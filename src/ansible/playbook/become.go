package playbook

import (
  //"ansible/playbook"
)

var become_fields = map[string]FieldAttribute{
}

type Become struct {
}

func (b *Become) Load(data map[interface{}]interface{}) {
  LoadValidFields(b, become_fields, data)
}
