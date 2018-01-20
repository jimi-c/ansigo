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
    field_name := "Attr_" + k
    field := s.FieldByName(field_name)
    if field_data, ok := data[strings.ToLower(k)]; ok {
      field.Set(reflect.ValueOf(field_data))
      delete(data, k)
    } else {
      if v.Default != nil {
        field.Set(reflect.ValueOf(v.Default))
      } else {
        field.Set(reflect.Zero(field.Type()))
      }
    }
  }
}

var base_fields = map[string]FieldAttribute{
  "name": FieldAttribute{T: "string", Default: ""},
}

type Base struct {
  squashed bool
  finalized bool

  Attr_name interface{}
}

func (b *Base) Name() string {
  name, _ := b.Attr_name.(string)
  return name
}

func (b *Base) Load(data map[interface{}]interface{}) {
  b.squashed = false
  b.finalized = false
  LoadValidFields(b, base_fields, data)
}

func TypeOf(v interface{}) string {
    switch t := v.(type) {
    case int:
        return "int"
    case float64:
        return "float64"
    case map[interface{}] interface{}:
        return "map"
    default:
        _ = t
        return "unknown"
    }
}
