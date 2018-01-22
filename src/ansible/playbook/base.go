package playbook

import (
  "fmt"
  "os"
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
  SkipLoad bool
}

// interfaces we use for playbooks

type Parent interface {
  GetInheritedValue(attr string) interface{}
}

type Validateable interface {
  GetAllObjectFieldAttributes() map[string]FieldAttribute
}

// Methods used for all Playbook structs, but not tied directly to them

func LoadValidFields(thing interface{}, field_map map[string]FieldAttribute, data map[interface{}]interface{}) {
  s := reflect.ValueOf(thing).Elem()
  for k, v := range field_map {
    if v.SkipLoad{
      // some special fields are contained in the FieldAttributes
      // for validation, but we don't want to load them here as we
      // load them specially in other ways
      continue
    }
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

func ValidateFields(thing Validateable, data map[interface{}]interface{}, is_task bool) {
  all_fields := thing.GetAllObjectFieldAttributes()
  for data_k, _ := range data {
    if ks, ok := data_k.(string); ok {
      found := false
      if _, found = all_fields[ks]; !found {
        if is_task {
          _, found = ModuleCache[ks]
        }
      }
      if !found {
        fmt.Println("Invalid field: ", ks)
        os.Exit(1)
      }
    } else {
      fmt.Println("Invalid field: ", data_k, "All fields must be string entries in YAML.")
      os.Exit(1)
    }
  }
}

func PostValidate(thing Validateable) Validateable {
  _ = thing.GetAllObjectFieldAttributes()
  return thing
}

// The base struct and related methods/etc.

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
