package playbook

import (
  "reflect"
)

var play_fields = map[string]FieldAttribute{
}

type Play struct {
  Base
  Become
  Taggable

  // role attributes
  //roles []Role
  // block and task lists are read from yaml, but not via
  // the normal LoadValidFields method.
  Handlers []Block
  Pre_tasks []Block
  Tasks []Block
  Post_tasks []Block

  // Field attributes read from yaml
  Attr_hosts interface{}
  // facts
  Attr_fact_path interface{}
  Attr_gather_facts interface{}
  Attr_gather_subset interface{}
  Attr_gather_timeout interface{}
  // variable attributes
  Attr_vars_files interface{}
  Attr_vars_prompt interface{}
  Attr_vault_password interface{}
  // flag/setting attributes
  Attr_force_handlers interface{}
  Attr_max_fail_percentage interface{}
  Attr_serial interface{}
  Attr_strategy interface{}
  Attr_order interface{}
}

func (p *Play) GetAllObjectFieldAttributes() map[string]FieldAttribute {
  var all_fields = make(map[string]FieldAttribute)
  var items = []map[string]FieldAttribute{base_fields, taggable_fields, become_fields, play_fields}
  for i := 0; i < len(items); i++ {
    for k, v := range items[i] {
      all_fields[k] = v
    }
  }
  return all_fields
}

func (p *Play) GetInheritedValue(attr string) interface{} {
  field_name := "Attr_" + attr
  s := reflect.ValueOf(p).Elem()
  field := s.FieldByName(field_name)

  var cur_value interface{}
  if field.Kind() != reflect.Invalid {
    cur_value = field.Interface()
  } else {
    cur_value = nil
  }
  return cur_value
}

func (p *Play) Load(data map[interface{}]interface{}) {
  p.Taggable.Load(data)
  p.Become.Load(data)

  LoadValidFields(p, play_fields, data)

  data_tasks, contains_tasks := data["tasks"]
  if contains_tasks {
    td, _ := data_tasks.([]interface{})
    p.Tasks = LoadListOfBlocks(td, p, p, false)
  }
}

func NewPlay(data map[interface{}]interface{}) *Play {
  p := new(Play)
  p.Load(data)
  return p
}
