package playbook

import (
)

var block_fields = map[string]FieldAttribute{
  "Delegate_To": FieldAttribute{
    T: "string", Default: "", Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Delegate_Facts": FieldAttribute{
    T: "bool", Default: false, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
}

type Block struct {
  Base
  Conditional
  Taggable
  Become

  // the parent object (a block, or another task)
  parent *Parent
  implicit_block bool
  block []interface{}
  rescue []interface{}
  always []interface{}

  // Field Attributes
  Delegate_To string
  Delegate_Facts bool
}

func (b *Block) GetInheritedValue() {

}

func (b *Block) Load(data map[interface{}]interface{}, play *Play, parent Parent, use_handlers bool) {
  b.Base.Load(data)
  b.Conditional.Load(data)
  b.Taggable.Load(data)
  b.Become.Load(data)

  LoadValidFields(b, block_fields, data)

  data_block, contains_block := data["block"]
  data_rescue, contains_rescue := data["rescue"]
  data_always, contains_always := data["always"]

  if contains_block {
    // FIXME handle errors here
    bb, _ := data_block.([]interface{})
    b.block = LoadListOfTasks(bb, play, b, use_handlers)
  } else {
    b.block = make([]interface{}, 0)
  }
  if contains_rescue {
    // FIXME handle errors here
    br, _ := data_rescue.([]interface{})
    b.rescue = LoadListOfTasks(br, play, b, use_handlers)
  } else {
    b.rescue = make([]interface{}, 0)
  }
  if contains_always {
    // FIXME handle errors here
    ba, _ := data_always.([]interface{})
    b.always = LoadListOfTasks(ba, play, b, use_handlers)
  } else {
    b.always = make([]interface{}, 0)
  }
}

func NewBlock(data map[interface{}]interface{}, play *Play, parent Parent, use_handlers bool) *Block {
  _, contains_block := data["block"]
  _, contains_rescue := data["rescue"]
  _, contains_always := data["always"]

  implicit := false
  if !(contains_block || contains_rescue || contains_always) {
    var data_list = []interface{}{data}
    data = map[interface{}]interface{} {
      "block": data_list,
    }
    implicit = true
  }

  b := new(Block)
  b.Load(data, play, parent, use_handlers)
  b.parent = &parent
  b.implicit_block = implicit

  return b
}
