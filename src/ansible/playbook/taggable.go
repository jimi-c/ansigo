package playbook

import (
)

var taggable_fields = map[string]FieldAttribute{
  "tags": FieldAttribute{T: "list", Default: nil},
}

type Taggable struct {
  Attr_tags interface{}
}

func (t *Taggable) Load(data map[interface{}]interface{}) {
  LoadValidFields(t, taggable_fields, data)
}
