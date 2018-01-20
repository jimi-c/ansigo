package playbook

import ()

var play_fields = map[string]FieldAttribute{
}

type Play struct {
  Taggable
  Become

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

func (p *Play) GetInheritedValue(attr string) interface{} {
  return nil
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
