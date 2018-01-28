package playbook

import (
  "fmt"
  "os"
  "path/filepath"
  "strings"
)

type ModuleInfo struct {
  Name string
  Path string
}

var ModuleCache = make(map[string]ModuleInfo)

func EnumerateModules(from_path string) {
  filepath.Walk(from_path, enumerateDirectory)
}

func enumerateDirectory(path string, info os.FileInfo, err error) error {
  if err != nil {
    return err
  }
  if info.IsDir() {
    filepath.Walk(path + info.Name(), enumerateDirectory)
  } else {
    // FIXME: error handling?
    abs_path, _ := filepath.Abs(path)
    ModuleCache[strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))] = ModuleInfo{info.Name(), abs_path}
  }
  return nil
}

func get_quote_state(token string, quote_char rune) rune {
  // the char before the current one, used to see if
  // the current character is escaped
  var prev_char byte = 0
  for idx, cur_char := range token {
    if idx > 0 { prev_char = token[idx-1] }
    if (cur_char == '"' || cur_char == '\'') && prev_char != '\\' {
      if quote_char != 0 {
        if cur_char == quote_char {
          quote_char = 0
        }
      } else {
        quote_char = cur_char
      }
    }
  }
  return quote_char
}

func count_jinja2_blocks(token string, cur_depth int, open_token string, close_token string) int {
  num_open := strings.Count(token, open_token)
  num_close := strings.Count(token, close_token)
  if num_open != num_close {
    cur_depth += (num_open - num_close)
    if cur_depth < 0 {
      cur_depth = 0
    }
  }
  return cur_depth
}

func SplitArgString(args string) []string {
  // the list of params parsed out of the arg string
  // this is going to be the result value when we are done
  params := make([]string, 0)
  args = strings.TrimSpace(args)
  items := strings.Split(args, "\n")

  // iterate over the tokens, and reassemble any that may have been
  // split on a space inside a jinja2 block.
  // ex if tokens are "{{", "foo", "}}" these go together

  // These variables are used to keep track of the state of the parsing,
  // since blocks and quotes may be nested within each other.

  var quote_char rune = 0
  var inside_quotes bool = false
  var print_depth int = 0    // used to count nested jinja2 {{ }} blocks
  var block_depth int = 0    // used to count nested jinja2 {% %} blocks
  var comment_depth int = 0  // used to count nested jinja2 {# #} blocks

  for itemidx, item := range items {
    item = strings.TrimSpace(item)
    tokens := strings.Split(item, " ")
    var line_continuation bool = false
    for idx, token := range tokens {
      if token == "\\" && !inside_quotes {
        line_continuation = true
        continue
      }

      was_inside_quotes := inside_quotes
      quote_char = get_quote_state(token, quote_char)
      inside_quotes = quote_char != 0

      var appended bool = false

      // if we're inside quotes now, but weren't before, append the token
      // to the end of the list, since we'll tack on more to it later
      // otherwise, if we're inside any jinja2 block, inside quotes, or we were
      // inside quotes (but aren't now) concat this token to the last param
      if inside_quotes && !was_inside_quotes && !(print_depth > 0 || block_depth > 0 || comment_depth > 0) {
        params = append(params, token)
        appended = true
      } else if print_depth > 0 ||  block_depth > 0 || comment_depth > 0 || inside_quotes || was_inside_quotes {
        end_idx := len(params) - 1
        if idx == 0 && was_inside_quotes {
          params[end_idx] = params[end_idx] + token
        } else if len(tokens) > 1 {
          var spacer string = ""
          if idx > 0 {
            spacer = " "
          }
          params[end_idx] = params[end_idx] + spacer + token
        } else {
          params[end_idx] = params[end_idx] + "\n" + token
        }
        appended = true
      }

      // if the number of paired block tags is not the same, the depth has changed, so we calculate that here
      // and may append the current token to the params (if we haven't previously done so)
      prev_print_depth := print_depth
      print_depth = count_jinja2_blocks(token, print_depth, "{{", "}}")
      if print_depth != prev_print_depth && !appended {
        params = append(params, token)
        appended = true
      }

      prev_block_depth := block_depth
      block_depth = count_jinja2_blocks(token, block_depth, "{%", "%}")
      if block_depth != prev_block_depth && !appended {
        params = append(params, token)
        appended = true
      }

      prev_comment_depth := comment_depth
      comment_depth = count_jinja2_blocks(token, comment_depth, "{#", "#}")
      if comment_depth != prev_comment_depth && !appended {
        params = append(params, token)
        appended = true
      }

      // finally, if we're at zero depth for all blocks and not inside quotes, and have not
      // yet appended anything to the list of params, we do so now
      if !(print_depth > 0 || block_depth > 0 || comment_depth > 0) && !inside_quotes && !appended && token != "" {
        params = append(params, token)
      }
    }

    // if this was the last token in the list, and we have more than
    // one item (meaning we split on newlines), add a newline back here
    // to preserve the original structure
    if len(items) > 1 && itemidx != len(items)-1 && !line_continuation {
      params[len(params)-1] += "\n"
    }

    // always clear the line continuation flag
    line_continuation = false
  }

  // If we're done and things are not at zero depth or we're still inside
  // quotes, raise an error to indicate that the args were unbalanced
  if print_depth != 0 || block_depth != 0 || comment_depth != 0 || inside_quotes {
    // FIXME
    //raise AnsibleParserError(u"failed at splitting arguments, either an unbalanced jinja2 block or quotes: {0}".format(args))
  }

  return params
}

func ParseKV(args string, check_raw bool) map[string]string {
  options := make(map[string]string)
  vargs := SplitArgString(args)

  raw_params := make([]string, 0)
  for _, orig_x := range vargs {
    pos := strings.Index(orig_x, "=")
    if pos > 0 && orig_x[pos-1] != '\\' {
      k := orig_x[:pos]
      v := orig_x[pos + 1:]
      _, is_special := map[string]string{"creates":"", "removes":"", "chdir":"", "executable":"", "warn":""}[k]
      if check_raw && is_special {
        raw_params = append(raw_params, orig_x)
      } else {
        options[k] = Unquote(strings.TrimSpace(v))
      }
    } else {
      raw_params = append(raw_params, orig_x)
    }
  }
  if len(raw_params) > 0 {
    options["_raw_params"] = strings.Join(raw_params, " ")
  }
  return options
}

func IsQuoted(data string) bool {
  return len(data) > 0 && data[0] == data[len(data)-1] && (data[0] == '"' || data[0] == '\'') && data[len(data)-2] != '\\'
}

func Unquote(data string) string {
  if IsQuoted(data) {
    return data[1:len(data)-1]
  }
  return data
}

func ExtendValue(cur_value []interface{}, new_value []interface{}, prepend bool) []interface{} {
  fmt.Println("* EXTENDING VALUES:", prepend)
  fmt.Println("  cur:", cur_value)
  fmt.Println("  new:", new_value)
  new_list := make([]interface{}, len(cur_value) + len(new_value))
  var one []interface{} = nil
  var two []interface{} = nil
  if prepend {
    one = new_value
    two = cur_value
  } else {
    one = cur_value
    two = new_value
  }
  last_i := 0
  for i, v := range one {
    new_list[i] = v
    last_i = i
  }
  for i, v := range two {
    new_list[last_i + i] = v
  }
  fmt.Println("- EXTENDED RESULT:", new_list)
  return new_list
}

func TypeOf(v interface{}) string {
    switch t := v.(type) {
    case int:
      return "int"
    case float64:
      return "float64"
    case map[interface{}] interface{}:
      return "map"
    case string:
      return "string"
    case interface{}:
      return "interface{}"
    default:
      _ = t
      return "unknown"
    }
}

func StringPos(value string, list []string) int {
  for p, v := range list {
    if (v == value) {
      return p
    }
  }
  return -1
}
