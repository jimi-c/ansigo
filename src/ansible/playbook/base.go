package playbook

import (
  "reflect"
  "strings"
)

type FieldAttribute struct {
  T string
  Default interface{}
  Required bool
  Priority int
  AlwaysPostValidate bool
  Inherit bool
  Alias []string
  Extend bool
  Prepend bool
}

/*
func NewFieldAttribute() *FieldAttribute {
  return FieldAttribute{
    Required: false,
    Priority: 0,
    AlwaysPostValidate: false,
    Inherit: true,
    Alias: make([]string, 0),
    Extend: false,
    Prepend: false,
  }
}
*/

func LoadValidFields(thing interface{}, field_map map[string]FieldAttribute, data map[interface{}]interface{}) {
  s := reflect.ValueOf(thing).Elem()
  for k, v := range field_map {
    field := s.FieldByName(k)
    if field_data, ok := data[strings.ToLower(k)]; ok {
      switch v.T {
      case "string":
        field.SetString(field_data.(string))
      case "map":
        _ = 1
      }
      delete(data, k)
    }
  }
}

var base_fields = map[string]FieldAttribute{
  "Name": FieldAttribute{T: "string", Default: ""},
}

type Base struct {
  Name string
}

func (b *Base) Load(data map[interface{}]interface{}) {
  LoadValidFields(b, base_fields, data)
}
