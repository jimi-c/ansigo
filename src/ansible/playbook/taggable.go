package playbook

import (
  //"ansible/playbook"
)

var taggable_fields = map[string]FieldAttribute{
}

type Taggable struct {
}

func (t *Taggable) Load(data map[interface{}]interface{}) {
  LoadValidFields(t, taggable_fields, data)
}
