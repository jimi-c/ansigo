package playbook

import (
)

var conditional_fields = map[string]FieldAttribute{
  "when": FieldAttribute{T: "list", Default: nil},
}

type Conditional struct {
  Attr_when interface{}
}

func (c *Conditional) Load(data map[interface{}]interface{}) {
  LoadValidFields(c, conditional_fields, data)
}
