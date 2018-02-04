package jinja2

import (
  "fmt"
  "errors"
  "unicode"
  "github.com/alecthomas/participle"
)

type TokenBoundary struct {
  Val string
  Pos int
  Line int
  NewLinePos int
}

type Renderable interface {
  Render(*Context) (string, error)
}

type DummyChunk struct {
  foo string
}
func (self *DummyChunk) Render(c *Context) (string, error) {
  return "", nil
}

type Token interface {
  //GetPos() int
}

type TokenBase struct {
  Pos, Line int
  StripBefore, StripAfter bool
}

type TextToken struct {
  TokenBase
  Text string
}

type TextChunk struct {
  Text string
}
func (self *TextChunk) Render(c *Context) (string, error) {
  return self.Text, nil
}

type VariableToken struct {
  TokenBase
  Content string
}

type VariableChunk struct {
  VarAst interface{}
  Content string
}
func (self *VariableChunk) Render(c *Context) (string, error) {
  //return self.VarAst.Eval(c), nil
  return "VARIABLE HERE", nil
}

type IfToken struct {
  TokenBase
  IfStatement string
}
type ElifToken struct {
  TokenBase
  ElifStatement string
}
type ElseToken struct {
  TokenBase
}
type EndifToken struct {
  TokenBase
}

type IfChunk struct {
  IfAst *IfStatement
  IfChunks []Renderable
  ElifChunks []ElifChunk
  ElseChunks []Renderable
}
func (self *IfChunk) Render(c *Context) (string, error) {
  chunk_rendered := false
  res := ""
  v, err := self.IfAst.Eval(c)
  if err != nil {
    // FIXME: error handling
    return "ERROR EVALUATING IF STATEMENT", err
  }
  v_bool, ok := v.Data.(bool)
  if !ok {
    return "COULD NOT CONVERT IF RETURN TO A BOOLEAN", nil
  }
  if v_bool {
    for _, chunk := range self.IfChunks {
      c_res, err := chunk.Render(c)
      if err != nil {
        return "", err
      } else {
        res = res + c_res
      }
    }
    chunk_rendered = true
  } else {
    for _, elif := range self.ElifChunks {
      v, err := elif.ElifAst.Eval(c)
      if err != nil {
        // FIXME: error handling
        return "ERROR EVALUATING ELIF STATEMENT", nil
      }
      v_bool, ok := v.Data.(bool)
      if !ok {
        return "COULD NOT CONVERT IF RETURN TO A BOOLEAN IN ELIF", nil
      }
      if v_bool {
        for _, chunk := range elif.ElifChunks {
          c_res, err := chunk.Render(c)
          if err != nil {
            return "", err
          } else {
            res = res + c_res
          }
        }
        chunk_rendered = true
        break
      }
    }
  }
  if !chunk_rendered {
    for _, chunk := range self.ElseChunks {
      c_res, err := chunk.Render(c)
      if err != nil {
        return "", err
      } else {
        res = res + c_res
      }
    }
  }
  return res, nil
}

type ElifChunk struct {
  ElifAst    *ElifStatement
  ElifChunks []Renderable
}

type ForToken struct {
  TokenBase
  ForStatement string // FIXME: should be a for expression
}

type ForChunk struct {
}
func (self *ForChunk) Render(c *Context) (string, error) { return "", nil }

type EndforToken struct {
  TokenBase
}

type RawToken struct {
  TokenBase
  Content string
}

type EndrawToken struct {
  TokenBase
}

type RawChunk struct {
  Content string
}
func (self *RawChunk) Render(c *Context) (string, error) {
  return self.Content, nil
}

func FindTokenBoundaries(input string) []TokenBoundary {
  //fmt.Println("TOKENIZING:")
  //fmt.Println(input)
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
        }
      } else if input[i] == '}' && i > 0 {
        sub := input[i-1:i+1]
        if sub == "%}" || sub == "}}" || sub == "#}" {
          token := TokenBoundary{sub, i, cur_line, cur_line_pos}
          bounds = append(bounds, token)
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

func ParseBlocks(tokens []Token, pos int, inside string) (int, []Renderable, error) {
  //fmt.Println("PARSING BLOCKS", pos)
  var contained_chunks []Renderable
  cur_pos := pos
  stop_parsing := false
  for cur_pos < len(tokens) {
    res := PeekToken(tokens[cur_pos])
    switch res {
    case "text":
      new_pos, text_chunk, err := ParseText(tokens, cur_pos)
      if err != nil {
        return cur_pos, nil, err
      }
      contained_chunks = append(contained_chunks, text_chunk)
      cur_pos = new_pos
    case "variable":
      new_pos, var_chunk, err := ParseVariable(tokens, cur_pos)
      if err != nil {
        return cur_pos, nil, err
      }
      contained_chunks = append(contained_chunks, var_chunk)
      cur_pos = new_pos
    case "if":
      new_pos, if_chunk, err := ParseIf(tokens, cur_pos)
      if err != nil {
        return cur_pos, nil, err
      }
      contained_chunks = append(contained_chunks, if_chunk)
      cur_pos = new_pos
    case "for":
      new_pos, for_chunk, err := ParseFor(tokens, cur_pos)
      if err != nil {
        return cur_pos, nil, err
      }
      contained_chunks = append(contained_chunks, for_chunk)
      cur_pos = new_pos
    case "elif", "endif":
      if inside == "if" {
        stop_parsing = true
      } else {
        return cur_pos, nil, errors.New("invalid token found: '" + res + "' but not currently inside an if statement")
      }
    case "endfor":
      if inside == "for" {
        stop_parsing = true
      } else {
        return cur_pos, nil, errors.New("invalid token found: '" + res + "' but not currently inside a for statement")
      }
    case "else":
      if inside == "for" || inside == "if" {
        stop_parsing = true
      } else {
        return cur_pos, nil, errors.New("invalid token found: '" + res + "' but not currently inside an if or a for statement")
      }
    case "raw":
      new_pos, raw_chunk, err := ParseRaw(tokens, cur_pos)
      if err != nil {
        return cur_pos, nil, err
      }
      contained_chunks = append(contained_chunks, raw_chunk)
      cur_pos = new_pos
    default:
      return cur_pos, nil, errors.New("invalid token found: '" + res + "'")
    }
    if stop_parsing { break }
  }
  //fmt.Println("DONE PARSING BLOCKS", cur_pos)
  return cur_pos, contained_chunks, nil
}
func ParseRaw(tokens []Token, pos int) (int, Renderable, error) {
  cur_pos := pos
  if res := PeekToken(tokens[cur_pos]); res != "raw" {
    return cur_pos, &DummyChunk{}, errors.New("expected a raw token, found '" + res + "' instead")
  }
  raw_token := tokens[cur_pos].(RawToken)
  cur_pos += 1
  if res := PeekToken(tokens[cur_pos]); res != "endraw" {
    return cur_pos, &DummyChunk{}, errors.New("expected a endraw token, found '" + res + "' instead")
  }
  raw_chunk := new(RawChunk)
  raw_chunk.Content = raw_token.Content
  return cur_pos+1, raw_chunk, nil
}
func ParseText(tokens []Token, pos int) (int, Renderable, error) {
  //fmt.Println("PARSING TEXT", pos)
  cur_pos := pos
  if res := PeekToken(tokens[cur_pos]); res != "text" {
    return cur_pos, &DummyChunk{}, errors.New("expected a text token, found '" + res + "' instead")
  } else {
    //fmt.Println("DONE PARSING TEXT")
    text_token := tokens[cur_pos].(TextToken)
    text_chunk := new(TextChunk)
    text_chunk.Text = text_token.Text
    return cur_pos+1, text_chunk, nil
  }
}
func ParseVariable(tokens []Token, pos int) (int, Renderable, error) {
  //fmt.Println("PARSING VARIABLE", pos)
  cur_pos := pos
  if res := PeekToken(tokens[cur_pos]); res != "variable" {
    return cur_pos, &DummyChunk{}, errors.New("expected a variable token, found '" + res + "' instead")
  } else {
    //fmt.Println("DONE PARSING VARIABLE")
    var_token := tokens[cur_pos].(VariableToken)
    var_chunk := new(VariableChunk)
    var_chunk.Content = var_token.Content
    return cur_pos+1, var_chunk, nil
  }
}

func ParseIfStatement(statement string) (*IfStatement, error) {
  parser, err := participle.Build(&IfStatement{}, nil)
  if err != nil {
    return nil, err
  }
  ast := &IfStatement{}
  if err := parser.ParseString(statement, ast); err != nil {
    return nil, err
  }
  return ast, nil
}
func ParseIf(tokens []Token, pos int) (int, Renderable, error) {
  //fmt.Println("PARSING IF", pos, "Number of Tokens:", len(tokens))
  if_chunk := new(IfChunk)
  if_chunk.IfAst = nil
  if_chunk.IfChunks = make([]Renderable, 0)
  if_chunk.ElifChunks = make([]ElifChunk, 0)
  if_chunk.ElseChunks = make([]Renderable, 0)

  cur_pos := pos
  if res := PeekToken(tokens[cur_pos]); res != "if" {
    return cur_pos, &DummyChunk{}, errors.New("expected an 'if' token but got '" + res + "' instead.")
  }

  if_token := tokens[cur_pos].(IfToken)
  ast, err := ParseIfStatement(if_token.IfStatement)
  if err != nil {
    return pos, &DummyChunk{}, err
  }
  if_chunk.IfAst = ast
  cur_pos += 1

  found_endif := false
  for cur_pos < len(tokens) {
    res := PeekToken(tokens[cur_pos])
    switch res {
    case "endif":
      found_endif = true
      cur_pos += 1
    case "elif":
      new_pos, elif_chunk, err := ParseElif(tokens, cur_pos)
      if err != nil {
        return cur_pos, &DummyChunk{}, err
      }
      if_chunk.ElifChunks = append(if_chunk.ElifChunks, elif_chunk)
      cur_pos = new_pos
    case "else":
      new_pos, else_chunks, err := ParseElse(tokens, cur_pos, "if")
      if err != nil {
        return cur_pos, &DummyChunk{}, err
      }
      if_chunk.ElseChunks = append(if_chunk.ElseChunks, else_chunks...)
      cur_pos = new_pos
    default:
      new_pos, contained_chunks, err := ParseBlocks(tokens, cur_pos, "if")
      if err != nil {
        return cur_pos, &DummyChunk{}, err
      }
      if_chunk.IfChunks = append(if_chunk.IfChunks, contained_chunks...)
      cur_pos = new_pos
    }
    if found_endif { break }
  }
  if !found_endif {
    //fmt.Println("NO ENDIF!!!!")
    return cur_pos, &DummyChunk{}, errors.New("Missing 'endif' for if statement tag.")
  }
  //fmt.Println("DONE PARSING IF STATEMENT")
  return cur_pos, if_chunk, nil
}

func ParseElifStatement(statement string) (*ElifStatement, error) {
  parser, err := participle.Build(&ElifStatement{}, nil)
  if err != nil {
    return nil, err
  }
  ast := &ElifStatement{}
  if err := parser.ParseString(statement, ast); err != nil {
    return nil, err
  }
  return ast, nil
}
func ParseElif(tokens []Token, pos int) (int, ElifChunk, error) {
  //fmt.Println("PARSING ELIF STATEMENT", pos)
  elif_chunk := new(ElifChunk)
  elif_chunk.ElifAst = nil
  elif_chunk.ElifChunks = make([]Renderable, 0)

  cur_pos := pos
  if res := PeekToken(tokens[cur_pos]); res != "elif" {
    return cur_pos, *elif_chunk, errors.New("expected an 'elif' token but got '" + res + "' instead.")
  }
  elif_token := tokens[cur_pos].(ElifToken)
  ast, err := ParseElifStatement(elif_token.ElifStatement)
  if err != nil {
    return pos, *elif_chunk, err
  }
  elif_chunk.ElifAst = ast
  cur_pos += 1

  found_stop := false
  for cur_pos < len(tokens) {
    res := PeekToken(tokens[cur_pos])
    switch res {
    case "elif", "else", "endif":
      found_stop = true
    default:
      new_pos, contained_chunks, err := ParseBlocks(tokens, cur_pos, "if")
      if err != nil {
        return cur_pos, *elif_chunk, err
      }
      elif_chunk.ElifChunks = append(elif_chunk.ElifChunks, contained_chunks...)
      cur_pos = new_pos
    }
    if found_stop { break }
  }
  //fmt.Println("DONE PARSING ELIF STATEMENT")
  return cur_pos, *elif_chunk, nil
}

func ParseElse(tokens []Token, pos int, in string) (int, []Renderable, error) {
  //fmt.Println("PARSING ELSE STATEMENT", pos)
  chunks := make([]Renderable, 0)
  cur_pos := pos
  if res := PeekToken(tokens[cur_pos]); res != "else" {
    return cur_pos, chunks, errors.New("expected an 'else' token but got '" + res + "' instead.")
  }
  cur_pos += 1

  found_stop := false
  for cur_pos < len(tokens) {
    res := PeekToken(tokens[cur_pos])
    switch res {
    case "endfor":
      if in == "for" {
        found_stop = true
      } else {
        return cur_pos, chunks, errors.New("unexpected 'endfor' found when not in a for loop.")
      }
    case "endif":
      if in == "if" {
        found_stop = true
      } else {
        return cur_pos, chunks, errors.New("unexpected 'endif' found when not in an if statement.")
      }
    default:
      new_pos, contained_chunks, err := ParseBlocks(tokens, cur_pos, in)
      if err != nil {
        return cur_pos, chunks, err
      }
      chunks = append(chunks, contained_chunks...)
      cur_pos = new_pos
    }
    if found_stop { break }
  }
  //fmt.Println("DONE PARSING ELSE STATEMENT")
  return cur_pos, chunks, nil
}
func ParseForStatement(statement string) (*ForStatement, error) {
  fmt.Println("PARSING FOR STATEMENT:", statement)
  parser, err := participle.Build(&ForStatement{}, nil)
  if err != nil {
    return nil, err
  }
  ast := &ForStatement{}
  if err := parser.ParseString(statement, ast); err != nil {
    return nil, err
  }
  fmt.Println("DONE PARSING FOR STATEMENT")
  return ast, nil
}
func ParseFor(tokens []Token, pos int) (int, Renderable, error) {
  //fmt.Println("PARSING FOR STATEMENT", pos)
  // For:
  // ForToken -> (TextToken|VariableToken|IfToken...|ForToken...) -> EndforToken
  cur_pos := pos
  var for_token ForToken
  //var contained_tokens []Token
  var endfor_token EndforToken

  if res := PeekToken(tokens[cur_pos]); res != "for" {
    return cur_pos, &DummyChunk{}, errors.New("expected a 'for' token but got '" + res + "' instead.")
  }
  for_token = tokens[cur_pos].(ForToken)
  cur_pos += 1

  ast, err := ParseForStatement(for_token.ForStatement)
  if err != nil {
    return pos, &DummyChunk{}, err
  }
  _ = ast

  found_endfor := false
  for cur_pos < len(tokens) {
    res := PeekToken(tokens[cur_pos])
    switch res {
    case "endfor":
      endfor_token = tokens[cur_pos].(EndforToken)
      found_endfor = true
      cur_pos += 1
      break
    case "else":
      new_pos, else_chunks, err := ParseElse(tokens, cur_pos, "for")
      if err != nil {
        return cur_pos, &DummyChunk{}, err
      }
      _ = else_chunks
      cur_pos = new_pos
    default:
      new_pos, contained_chunks, err := ParseBlocks(tokens, cur_pos, "for")
      if err != nil {
        return cur_pos, &DummyChunk{}, err
      }
      _ = contained_chunks
      cur_pos = new_pos
    }
  }
  if !found_endfor {
    return cur_pos, &DummyChunk{}, errors.New("Missing 'endfor' for for loop tags.")
  }
  _ = for_token
  _ = endfor_token
  //fmt.Println("DONE PARSING FOR STATEMENT")
  return cur_pos, &DummyChunk{}, nil
}
