package playbook

import (
)

var become_fields = map[string]FieldAttribute{
}

type Become struct {
  // methods we override from the top-level composed class
  GetInheritedValue func (string) interface{}
  GetAllObjectFieldAttributes func() map[string]FieldAttribute
}

func (b *Become) Load(data map[interface{}]interface{}) {
  LoadValidFields(b, become_fields, data)
}
