package playbook

import (
  "reflect"
)

var task_fields = map[string]FieldAttribute{
  "async_val": FieldAttribute{
    T: "int", Default: 0, Required: false, Priority: 0, Inherit: true, Alias: []string{"async"}, Extend: false, Prepend: false,
  },
  "changed_when": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  "delay": FieldAttribute{
    T: "int", Default: 5, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "delegate_to": FieldAttribute{
    T: "string", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "delegate_facts": FieldAttribute{
    T: "bool", Default: false, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "failed_when": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  "loop": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  //"Loop_Control": FieldAttribute{
  //  T: "struct", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{"async"}, Extend: false, Prepend: false,
  //},
  "notify": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  "poll": FieldAttribute{
    T: "int", Default: 10, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "register": FieldAttribute{
    T: "string", Default: "", Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "retries": FieldAttribute{
    T: "int", Default: 3, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "until": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
}

type Task struct {
  // composed structs
  Base
  Become
  Conditional
  Taggable

  // the parent object (a block, or another task)
  parent Parent

  Attr_action interface{}
  Attr_args interface{}

  // Field Attributes
  Attr_async_val interface{}
  Attr_changed_when interface{}
  Attr_delay interface{}
  Attr_delegate_to interface{}
  Attr_delegate_facts interface{}
  Attr_failed_when interface{}
  Attr_loop interface{}
  //loop_control interface{}
  Attr_notify interface{}
  Attr_poll interface{}
  Attr_register interface{}
  Attr_retries interface{}
  Attr_until interface{}
}

func (t *Task) GetAllObjectFieldAttributes() map[string]FieldAttribute {
  var all_fields = make(map[string]FieldAttribute)
  var items = []map[string]FieldAttribute{base_fields, conditional_fields, taggable_fields, become_fields, task_fields}
  for i := 0; i < len(items); i++ {
    for k, v := range items[i] {
      all_fields[k] = v
    }
  }
  return all_fields
}

func (t *Task) GetInheritedValue(attr string) interface{} {
  all_fields := t.GetAllObjectFieldAttributes()
  field_attribute := all_fields[attr]

  field_name := "Attr_" + attr
  s := reflect.ValueOf(t).Elem()
  field := s.FieldByName(field_name)

  var cur_value interface{}
  if field.Kind() != reflect.Invalid {
    cur_value = field.Interface()
  } else {
    cur_value = nil
  }

  get_parent_value := field_attribute.Inherit &&
                      t.parent != nil &&
                      cur_value != reflect.Zero(field.Type()) &&
                      !(t.squashed || t.finalized)
  // FIXME: do append and prepend stuff here too, as well as other
  //        considerations from the python version such as dynamic
  //        includes, etc.
  if get_parent_value || field_attribute.Extend {
    parent_value := t.parent.GetInheritedValue(attr)
    if parent_value != reflect.Zero(field.Type()) && parent_value != nil {
      if field_attribute.Extend && cur_value != nil {
        cur_value = ExtendValue(cur_value.([]interface{}), parent_value.([]interface{}), field_attribute.Prepend)
      } else {
        cur_value = parent_value
      }
    }
  }

  return cur_value
}

func (t *Task) Load(data map[interface{}]interface{}) {
  t.Base.Load(data)
  t.Conditional.Load(data)
  t.Taggable.Load(data)
  t.Become.Load(data)

  LoadValidFields(t, task_fields, data)

  t.Base.GetInheritedValue = t.GetInheritedValue
  t.Base.GetAllObjectFieldAttributes = t.GetAllObjectFieldAttributes
  t.Conditional.GetInheritedValue = t.GetInheritedValue
  t.Conditional.GetAllObjectFieldAttributes = t.GetAllObjectFieldAttributes
  t.Taggable.GetInheritedValue = t.GetInheritedValue
  t.Taggable.GetAllObjectFieldAttributes = t.GetAllObjectFieldAttributes
  t.Become.GetInheritedValue = t.GetInheritedValue
  t.Become.GetAllObjectFieldAttributes = t.GetAllObjectFieldAttributes

  for k, v := range data {
    if _, ok := ModuleCache[k.(string)]; ok || k.(string) == "setup" {
      t.Attr_action = k.(string)
      switch s := TypeOf(v); s {
        case "map":
          args := make(map[string]interface{})
          if arg_map, ok := v.(map[interface{}]interface{}); ok {
            for k, v := range arg_map {
              if str_k, ok := k.(string); ok {
                args[str_k] = v
              } else {
                // FIXME: error handling
              }
            }
          } else {
            // FIXME: error handling
          }
          t.Attr_args = args
        case "string":
          raw_modules := map[string]string{"command":"", "shell":"", "script":""}
          _, check_raw := raw_modules[k.(string)]
          t.Attr_args = ParseKV(v.(string), check_raw)
        default:
          t.Attr_args = make(map[string]interface{})
      }
      delete(data, k.(string))
    }
  }
}

func (t *Task) EvaluateTags(only_tags []string, skip_tags []string) bool {
  return EvaluateTags(t, only_tags, skip_tags)
}

func (t *Task) EvaluateConditional() bool {
  return EvaluateConditional(t)
}

// local getters
func (t *Task) Action() string {
  if res, ok := t.Attr_action.(string); ok {
    return res
  }
  return ""
}
func (t *Task) Args() map[string]interface{} {
  args := make(map[string]interface{})
  if res, ok := t.Attr_args.(map[string]interface{}); ok {
    args = res
  } else {
    // FIXME: error handling
  }
  return args
}
func (t *Task) AsyncVal() int {
  if res, ok := t.GetInheritedValue("async_val").(int); ok {
    return res
  } else {
    res, _ := task_fields["async_val"].Default.(int)
    return res
  }
}
func (t *Task) ChangedWhen() []string {
  if res, ok := t.GetInheritedValue("changed_when").([]string); ok {
    return res
  } else {
    res, _ := task_fields["changed_when"].Default.([]string)
    return res
  }
}
func (t *Task) Delay() int {
  if res, ok := t.GetInheritedValue("delay").(int); ok {
    return res
  } else {
    res, _ := task_fields["delay"].Default.(int)
    return res
  }
}
func (t *Task) DelegateTo() string {
  if res, ok := t.GetInheritedValue("delegate_to").(string); ok {
    return res
  } else {
    res, _ := task_fields["delegate_to"].Default.(string)
    return res
  }
}
func (t *Task) DelegateFacts() bool {
  if res, ok := t.GetInheritedValue("delegate_facts").(bool); ok {
    return res
  } else {
    res, _ := task_fields["delegate_facts"].Default.(bool)
    return res
  }
}
func (t *Task) FailedWhen() []string {
  if res, ok := t.GetInheritedValue("failed_when").([]string); ok {
    return res
  } else {
    res, _ := task_fields["failed_when"].Default.([]string)
    return res
  }
}
func (t *Task) Loop() []string {
  if res, ok := t.GetInheritedValue("loop").([]string); ok {
    return res
  } else {
    res, _ := task_fields["loop"].Default.([]string)
    return res
  }
}
func (t *Task) Notify() []string {
  if res, ok := t.GetInheritedValue("notify").([]string); ok {
    return res
  } else {
    res, _ := task_fields["notify"].Default.([]string)
    return res
  }
}
func (t *Task) Poll() int {
  if res, ok := t.GetInheritedValue("poll").(int); ok {
    return res
  } else {
    res, _ := task_fields["poll"].Default.(int)
    return res
  }
}
func (t *Task) Retries() int {
  if res, ok := t.GetInheritedValue("retries").(int); ok {
    return res
  } else {
    res, _ := task_fields["retries"].Default.(int)
    return res
  }
}
func (t *Task) Register() string {
  if res, ok := t.GetInheritedValue("register").(string); ok {
    return res
  } else {
    res, _ := task_fields["register"].Default.(string)
    return res
  }
}
func (t *Task) Until() []string {
  if res, ok := t.GetInheritedValue("until").([]string); ok {
    return res
  } else {
    res, _ := task_fields["until"].Default.([]string)
    return res
  }
}

// the generator function for tasks
func NewTask(data map[interface{}]interface{}, parent Parent) *Task {
  t := new(Task)
  ValidateFields(t, data, true)
  t.parent = parent
  t.Load(data)
  return t
}
