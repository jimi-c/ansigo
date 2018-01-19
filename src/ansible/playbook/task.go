package playbook

import (
)

type Parent interface {
  GetInheritedValue()
}

var task_fields = map[string]FieldAttribute{
  "Async_Val": FieldAttribute{
    T: "int", Default: 0, Required: false, Priority: 0, Inherit: true, Alias: []string{"async"}, Extend: false, Prepend: false,
  },
  "Changed_When": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  "Delay": FieldAttribute{
    T: "int", Default: 5, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Delegate_To": FieldAttribute{
    T: "string", Default: "", Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Delegate_Facts": FieldAttribute{
    T: "bool", Default: false, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Failed_When": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  "Loop": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  //"Loop_Control": FieldAttribute{
  //  T: "struct", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{"async"}, Extend: false, Prepend: false,
  //},
  "Notify": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
  "Poll": FieldAttribute{
    T: "int", Default: 10, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Register": FieldAttribute{
    T: "string", Default: "", Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Retries": FieldAttribute{
    T: "int", Default: 3, Required: false, Priority: 0, Inherit: true, Alias: make([]string, 0), Extend: false, Prepend: false,
  },
  "Until": FieldAttribute{
    T: "list", Default: nil, Required: false, Priority: 0, Inherit: true, Alias: []string{}, Extend: false, Prepend: false,
  },
}

type Task struct {
  // composed structs
  Base
  Conditional
  Taggable
  Become

  // the parent object (a block, or another task)
  parent *Parent

  Action string
  Args map[interface{}]interface{}

  // Field Attributes
  Async_Val int
  Changed_When []string
  Delay int
  Delegate_To string
  Delegate_Facts bool
  Failed_When []string
  Loop []interface{}
  //Loop_Control LoopControl
  Notify []string
  Poll int
  Register string
  Retries int
  Until []string
}

func (t *Task) GetInheritedValue() {
}

func (t *Task) Load(data map[interface{}]interface{}) {
  t.Base.Load(data)
  t.Conditional.Load(data)
  t.Taggable.Load(data)
  t.Become.Load(data)

  LoadValidFields(t, task_fields, data)

  for k, v := range data {
    if k.(string) == "debug" {
      t.Action = k.(string)
      switch s := TypeOf(v); s {
        case "map":
          t.Args = v.(map[interface{}] interface{})
        default:
          t.Args = make(map[interface{}] interface{})
      }
      delete(data, k.(string))
    }
  }
}

func NewTask(data map[interface{}]interface{}, parent Parent) *Task {
  t := new(Task)
  t.parent = &parent
  t.Load(data)
  return t
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
