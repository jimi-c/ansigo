package playbook

import (
  "reflect"
)

var block_fields = map[string]FieldAttribute{
  "block": FieldAttribute{SkipLoad: true},
  "rescue": FieldAttribute{SkipLoad: true},
  "always": FieldAttribute{SkipLoad: true},
  "delegate_to": FieldAttribute{
    T: "string", Default: "", Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "delegate_facts": FieldAttribute{
    T: "bool", Default: false, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
}

type Block struct {
  Base
  Become
  Conditional
  Taggable

  // the parent object (a block, or another task)
  parent *Parent
  implicit_block bool

  // read from yaml, but loaded recursively by helpers
  Attr_block []interface{}
  Attr_rescue []interface{}
  Attr_always []interface{}

  // attributes read from yaml directly
  Attr_delegate_to interface{}
  Attr_delegate_facts interface{}
}

func (b *Block) GetAllObjectFieldAttributes() map[string]FieldAttribute {
  var all_fields = make(map[string]FieldAttribute)
  var items = []map[string]FieldAttribute{base_fields, conditional_fields, taggable_fields, become_fields, block_fields}
  for i := 0; i < len(items); i++ {
    for k, v := range items[i] {
      all_fields[k] = v
    }
  }
  return all_fields
}

func (b *Block) GetInheritedValue(attr string) interface{} {
  all_fields := b.GetAllObjectFieldAttributes()
  field_attribute := all_fields[attr]

  field_name := "Attr_" + attr
  s := reflect.ValueOf(b).Elem()
  field := s.FieldByName(field_name)

  var cur_value interface{}
  if field.Kind() != reflect.Invalid {
    cur_value = field.Interface()
  } else {
    cur_value = nil
  }

  get_parent_value := field_attribute.Inherit &&
                      b.parent != nil &&
                      cur_value != reflect.Zero(field.Type()) &&
                      !(b.squashed || b.finalized)
  if get_parent_value {
    parent_value := (*b.parent).GetInheritedValue(attr)
    if parent_value != reflect.Zero(field.Type()) && parent_value != nil {
      cur_value = parent_value
    }
  }

  return cur_value
}

func (b *Block) Load(data map[interface{}]interface{}, play *Play, parent Parent, use_handlers bool) {
  b.Base.Load(data)
  b.Conditional.Load(data)
  b.Taggable.Load(data)
  b.Become.Load(data)

  data_block, contains_block := data["block"]
  data_rescue, contains_rescue := data["rescue"]
  data_always, contains_always := data["always"]

  if contains_block {
    // FIXME handle errors here
    bb, _ := data_block.([]interface{})
    b.Attr_block = LoadListOfTasks(bb, play, b, use_handlers)
    delete(data, "block")
  } else {
    b.Attr_block = make([]interface{}, 0)
  }
  if contains_rescue {
    // FIXME handle errors here
    br, _ := data_rescue.([]interface{})
    b.Attr_rescue = LoadListOfTasks(br, play, b, use_handlers)
    delete(data, "rescue")
  } else {
    b.Attr_rescue = make([]interface{}, 0)
  }
  if contains_always {
    // FIXME handle errors here
    ba, _ := data_always.([]interface{})
    b.Attr_always = LoadListOfTasks(ba, play, b, use_handlers)
    delete(data, "always")
  } else {
    b.Attr_always = make([]interface{}, 0)
  }

  LoadValidFields(b, block_fields, data)
}

func (b *Block) Copy() *Block {
  new_block := new(Block)
  new_block.parent = b.parent
  new_block.implicit_block = b.implicit_block
  old_s := reflect.ValueOf(b).Elem()
  new_s := reflect.ValueOf(new_block).Elem()
  for k, _ := range b.GetAllObjectFieldAttributes() {
    field_name := "Attr_" + k
    old_field := old_s.FieldByName(field_name)
    new_field := new_s.FieldByName(field_name)
    new_field.Set(old_field)
  }
  new_block.Attr_block = make([]interface{}, len(b.Attr_block))
  copy(new_block.Attr_block, b.Attr_block)
  new_block.Attr_rescue = make([]interface{}, len(b.Attr_rescue))
  copy(new_block.Attr_rescue, b.Attr_rescue)
  new_block.Attr_always = make([]interface{}, len(b.Attr_always))
  copy(new_block.Attr_always, b.Attr_always)
  return new_block
}

// local getters
func (b *Block) DelegateTo() string {
  if res, ok := b.GetInheritedValue("delegate_to").(string); ok {
    return res
  } else {
    res, _ := block_fields["delegate_to"].Default.(string)
    return res
  }
}
func (b *Block) DelegateFacts() bool {
  if res, ok := b.GetInheritedValue("delegate_facts").(bool); ok {
    return res
  } else {
    res, _ := block_fields["delegate_facts"].Default.(bool)
    return res
  }
}
// base mixin getters
// become mixin getters
// conditional mixin getters
func (b *Block) When() []string {
  if res, ok := b.GetInheritedValue("when").([]string); ok {
    return res
    } else {
      res, _ := conditional_fields["when"].Default.([]string)
      return res
    }
  }
// taggable mixin getters
func (b *Block) Tags() []string {
  if res, ok := b.GetInheritedValue("tags").([]string); ok {
    return res
  } else {
    res, _ := taggable_fields["tags"].Default.([]string)
    return res
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
  ValidateFields(b, data, false)
  b.Load(data, play, parent, use_handlers)
  b.parent = &parent
  b.implicit_block = implicit

  return b
}
