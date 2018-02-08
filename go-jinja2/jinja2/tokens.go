package jinja2

import (
  //"fmt"
)

type TokenBoundary struct {
  Val string
  Pos int
  Line int
  NewLinePos int
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

type VariableToken struct {
  TokenBase
  Content string
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

type ForToken struct {
  TokenBase
  ForStatement string
}

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
