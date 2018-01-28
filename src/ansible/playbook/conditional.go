package playbook

import (
)

type ConditionalEvaluate interface {
  EvaluateConditional() bool
}

var conditional_fields = map[string]FieldAttribute{
  "when": FieldAttribute{T: "list", Default: nil, Extend: true, Prepend: true},
}

type Conditional struct {
  Attr_when interface{}
}

func (c *Conditional) Load(data map[interface{}]interface{}) {
  LoadValidFields(c, conditional_fields, data)
}

func EvaluateConditional(thing ConditionalEvaluate) bool {
  return true
}
