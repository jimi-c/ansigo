package jinja2

import (
  "errors"
  "unicode"
)

type Template struct {
  data string
  template_chunks []Renderable
}

func (self *Template) Parse(data string) error {
  self.data = data
  self.template_chunks = make([]Renderable, 0)

  tokens := Tokenize(self.data)
  for pos := 0; pos < len(tokens); {
    new_pos, contained_chunks, err := ParseBlocks(tokens, pos, "")
    if err != nil {
      return err
    }
    self.template_chunks = append(self.template_chunks, contained_chunks...)
    pos = new_pos
  }
  return nil
}

func (self *Template) Render(c *Context) (string, error) {
  res := ""
  for _, chunk := range self.template_chunks {
    c_res, err := chunk.Render(c)
    if err != nil {
      return "", err
    } else {
      res = res + c_res
    }
  }
  return res, nil
}

func FindTokenBoundaries(input string) []TokenBoundary {
  bounds := make([]TokenBoundary, 0)
  cur_line := 0
  cur_line_pos := 0
  in_quotes := false
  quote_char := byte(0)
  for i := 0; i < len(input); i++ {
    if in_quotes {
      if input[i] == quote_char {
        in_quotes = false
        quote_char = byte(0)
      }
    } else if input[i] == '\n' {
      cur_line++
    } else {
      if input[i] == '\'' || input[i] == '"' {
        in_quotes = true
        quote_char = input[i]
      } else if input[i] == '{' && i < len(input) {
        sub := input[i:i+2]
        if sub == "{%" || sub == "{{" || sub == "{#" {
          token := TokenBoundary{sub, i, cur_line, cur_line_pos}
          bounds = append(bounds, token)
          i += 1
        }
      } else if input[i] == '}' && i > 0 {
        sub := input[i-1:i+1]
        if sub == "%}" || sub == "}}" || sub == "#}" {
          if sub != "}}" || i >= len(input) - 1 || input[i+1] != '}' {
            token := TokenBoundary{sub, i, cur_line, cur_line_pos}
            bounds = append(bounds, token)
          }
        }
      }
    }
  }
  return bounds
}

func GetNextId(input string, start int) (string, int, error) {
  consuming := false
  found := make([]rune, 0)
  for i := start; i < len(input); i++ {
    cur_c := rune(input[i])
    if !unicode.IsLetter(cur_c) {
      if consuming {
        return string(found), i, nil
      }
    } else {
      if !consuming {
        consuming = true
      }
      found = append(found, cur_c)
    }
  }
  if string(found) == "" {
    return "", start, errors.New("no id found")
  } else {
    return string(found), len(input), nil
  }
}

func Tokenize(input string) []Token {
  res := FindTokenBoundaries(input)
  cur_pos := 0
  end_pos := 0
  cur_line := 1
  cur_line_pos := 0
  tokens := make([]Token, 0)
  in_raw := false
  for i := 0; i < len(res); {
    t1 := res[i]
    t2 := res[i+1]
    if in_raw {
      if t1.Val != "{%" && t2.Val != "%}" {
        i += 1
        continue
      }
      strip_before := false
      strip_after := false
      block_statement := input[t1.Pos+2:t2.Pos-1]
      if len(block_statement) > 0 && block_statement[0] == '-' {
        block_statement = block_statement[1:]
        strip_before = true
      }
      if len(block_statement) > 0 && block_statement[len(block_statement)-1] == '-' {
        block_statement = block_statement[:len(block_statement)-1]
        strip_after = true
      }
      id, idpos, err := GetNextId(block_statement, 0)
      if err != nil {
        panic(err)
      } else {
        if id != "endraw" {
          i += 1
          continue
        }
        next_id, _, _ := GetNextId(block_statement, idpos)
        if next_id != "" {
          panic("endraw statements can't have any thing else with them")
        }
        end_pos = t1.Pos
        if end_pos > cur_pos {
          raw_token := tokens[len(tokens)-1].(RawToken)
          raw_token.Content = raw_token.Content + input[cur_pos:end_pos]
          tokens[len(tokens)-1] = raw_token
        }
        endraw_token := EndrawToken{TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
        tokens = append(tokens, endraw_token)
        in_raw = false
        cur_pos = t2.Pos + 1
        i += 2
      }
    } else {
      if t1.Val == "{{" && t2.Val != "}}" || t1.Val == "{%" && t2.Val != "%}" || t1.Val == "{#" && t2.Val != "#}" {
        panic("mismatched boundaries: " + t2.Val + " was found immediately after " + t1.Val)
      }
      end_pos = t1.Pos
      if end_pos > cur_pos {
        // text block
        text_token := TextToken{Text: input[cur_pos:end_pos], TokenBase: TokenBase{t1.Pos, t1.Line, false, false}}
        tokens = append(tokens, text_token)
      }
      cur_line = t2.Line
      cur_line_pos = t2.NewLinePos
      if t1.Val == "{{" {
        // variable block
        var_data := input[t1.Pos+2:t2.Pos-1]
        var_token := VariableToken{Content: var_data, TokenBase: TokenBase{t1.Pos, t1.Line, false, false}}
        tokens = append(tokens, var_token)
      } else if t1.Val == "{%" {
        // if/for/something block
        strip_before := false
        strip_after := false
        block_statement := input[t1.Pos+2:t2.Pos-1]
        if len(block_statement) > 0 && block_statement[0] == '-' {
          block_statement = block_statement[1:]
          strip_before = true
        }
        if len(block_statement) > 0 && block_statement[len(block_statement)-1] == '-' {
          block_statement = block_statement[:len(block_statement)-1]
          strip_after = true
        }
        id, idpos, err := GetNextId(block_statement, 0)
        if err != nil {
          panic(err)
        } else {
          var token_thing Token
          switch id {
          case "if":
            token_thing = IfToken{IfStatement: block_statement, TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
          case "elif":
            token_thing = ElifToken{ElifStatement: block_statement, TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
          case "else":
            next_id, _, _ := GetNextId(block_statement, idpos)
            if next_id != "" {
              panic("else statements can't have any thing else with them")
            }
            //fmt.Println("CREATING ELSE TOKEN FROM", block_statement)
            token_thing = ElseToken{TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
          case "endif":
            next_id, _, _ := GetNextId(block_statement, idpos)
            if next_id != "" {
              panic("endif statements can't have any thing else with them")
            }
            token_thing = EndifToken{TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
          case "for":
            token_thing = ForToken{ForStatement: block_statement, TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
          case "endfor":
            next_id, _, _ := GetNextId(block_statement, idpos)
            if next_id != "" {
              panic("endfor statements can't have any thing else with them")
            }
            token_thing = EndforToken{TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
          case "raw":
            next_id, _, _ := GetNextId(block_statement, idpos)
            if next_id != "" {
              panic("raw statements can't have any thing else with them")
            }
            token_thing = RawToken{Content: "", TokenBase: TokenBase{t1.Pos+1, t1.Line+1, strip_before, strip_after}}
            in_raw = true
          default:
            panic("invalid logical id:'" + id + "'")
          }
          tokens = append(tokens, token_thing)
        }
      } else if t1.Val == "{#" {
        // comment block
        //fmt.Println("COMMENT BLOCK: '" + input[t1.Pos:t2.Pos+1] + "'")
      }
      cur_pos = t2.Pos + 1
      i += 2
    }
  }
  if cur_pos < len(input) {
    text_token := TextToken{Text: input[cur_pos:], TokenBase: TokenBase{cur_pos - cur_line_pos + 1, cur_line+1, false, false}}
    tokens = append(tokens, text_token)
  }
  return tokens
}

func PeekToken(token Token) string {
  switch token.(type) {
  case TextToken:
    return "text"
  case VariableToken:
    return "variable"
  case IfToken:
    return "if"
  case ElifToken:
    return "elif"
  case ElseToken:
    return "else"
  case EndifToken:
    return "endif"
  case ForToken:
    return "for"
  case EndforToken:
    return "endfor"
  case RawToken:
    return "raw"
  case EndrawToken:
    return "endraw"
  }
  return "unknown"
}
