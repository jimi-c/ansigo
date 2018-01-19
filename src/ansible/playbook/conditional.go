package playbook

import (
  //"ansible/playbook"
)

var conditional_fields = map[string]FieldAttribute{
  "When": FieldAttribute{T: "string"},
}

type Conditional struct {
  When string
}

func (c *Conditional) Load(data map[interface{}]interface{}) {
  LoadValidFields(c, conditional_fields, data)
}
