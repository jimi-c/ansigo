package playbook

import (
  "errors"
  "github.com/jimi-c/jinja2"
)

type ConditionalEvaluate interface {
  EvaluateConditional() bool
  When() []string
}

var conditional_fields = map[string]FieldAttribute{
  "when": FieldAttribute{T: "list", Default: nil, Extend: true, Prepend: true},
}

type Conditional struct {
  Attr_when interface{}
  // methods we override from the top-level composed class
  GetInheritedValue func (string) interface{}
  GetAllObjectFieldAttributes func() map[string]FieldAttribute
}

func (c *Conditional) Load(data map[interface{}]interface{}) {
  LoadValidFields(c, conditional_fields, data)
}

// conditional mixin getters
func (c *Conditional) When() []string {
  if res, ok := c.GetInheritedValue("when").([]string); ok {
    return res
  } else {
    res, _ := conditional_fields["when"].Default.([]string)
    return res
  }
}

func EvaluateConditional(thing ConditionalEvaluate) (bool, error) {
  context := jinja2.NewContext(nil)
  context.AddVariables(map[string]interface{} {
      "foo": true,
    },
  )
  template := new(jinja2.Template)
  for _, cond := range thing.When() {
    err := template.Parse(`{% if ` + cond + ` %}True{% else %}False{% endif %}`)
    if err != nil {
      return false, err
    }
    if res, err := template.Render(context); err != nil {
      return false, err
    } else {
      if res == "True" {
        return true, nil
      } else if res == "False" {
        return false, nil
      } else {
        return false, errors.New("Unknown result returned from conditional statement evaluation.")
      }
    }
  }
  return true, nil
}
