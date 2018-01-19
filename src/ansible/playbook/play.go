package playbook

import ()

var play_fields = map[string]FieldAttribute{
}

type Play struct {
  Taggable
  Become

  // Field Attributes
  hosts []string
  // facts
  fact_path string
  gather_facts bool
  gather_subset []string
  gather_timeout int
  // variable attributes
  vars_files []string
  vars_prompt []interface{}
  vault_password string
  // role attributes
  //roles []Role
  // block and task lists
  handlers []Block
  pre_tasks []Block
  tasks []Block
  post_tasks []Block
  // flag/setting attributes
  force_handlers bool
  max_fail_percentage float64
  serial []int
  strategy string
  order string
}

func (p *Play) GetInheritedValue() {
}

func (p *Play) Load(data map[interface{}]interface{}) {
  p.Taggable.Load(data)
  p.Become.Load(data)

  LoadValidFields(p, play_fields, data)

  data_tasks, contains_tasks := data["tasks"]
  if contains_tasks {
    td, _ := data_tasks.([]interface{})
    p.tasks = LoadListOfBlocks(td, p, p, false)
  }
}

func NewPlay(data map[interface{}]interface{}) *Play {
  p := new(Play)
  p.Load(data)
  return p
}
