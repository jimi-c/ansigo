package gojinja2

import ()

var Filters = new(map[string]func())

type Block interface {
  Render() string
  HasChildren() bool
  GetChildren() []Block
}

type TextBlock struct {
  Text string
}
func (self *TextBlock) Render() {
  return ""
}
func (self *TextBlock) HasChildren() bool {
  return false
}
func (self *TextBlock) GetChildren() []Block {
  // text blocks never have children, so we return nil here
  return nil
}

type Token struct {
  Val string
  Pos int
}

func TokenizeInput(input string) []Token {
  tokens := make([]Token, 0)
  for i := 0; i < len(input); i++ {
    if input[i] == '{' && i < len(input) {
      sub := input[i:i+1]
      if sub == "{%" || sub == "{{" || sub == "{#" {
        token = Token{sub, i}
      }
    } else if input[i] == '}' && i > 0 {
      sub := input[i-1:i]
      if sub == "{%" || sub == "{{" || sub == "{#" {
        token = Token{sub, i}
      }
    }
  }
  return tokens
}

func ParseText(input string, start int) (int, int, err) {
  // read input from start until we either hit the end
  // of the string or until we find a block/var token
  return start, start, nil
}

func ParseVariable(input string, start int) (int, int, err) {
  // read from {{ to }}, which may contain 0 or more |filter expressions
  // if we run into any blocks here, we error
  return start, start, nil
}

func ParseIfBlock(input string, start int) (int, int, err) {
  // we should read:
  // {% if <statement> %}
  // <block>
  // [{% elif %}<block>]*
  // [{% else %}<block]
  // {% endif %}
  return start, start, nil
}
