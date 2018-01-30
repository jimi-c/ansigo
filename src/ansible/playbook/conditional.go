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
  // methods we override from the top-level composed class
  GetInheritedValue func (string) interface{}
  GetAllObjectFieldAttributes func() map[string]FieldAttribute
}

func (c *Conditional) Load(data map[interface{}]interface{}) {
  LoadValidFields(c, conditional_fields, data)
}

// conditional mixin getters
func (c *Conditional) When() []string {
  if res, ok := c.GetInheritedValue("when").([]string); ok {
    return res
  } else {
    res, _ := conditional_fields["when"].Default.([]string)
    return res
  }
}

func EvaluateConditional(thing ConditionalEvaluate) bool {
  return true
}
