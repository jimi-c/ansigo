package jinja2

import (
  "strconv"
)

type Context struct {
  Variables map[string]VariableType
  Filters map[string]func(VariableType, ...interface{})(VariableType, error)
  Tests map[string]func(VariableType, ...interface{})(VariableType, error)
}
func (self *Context) LoadDefaultFilters() {
  self.Filters["int"] = func(val VariableType, args...interface{}) (VariableType, error) {
    switch val.Type {
    case PY_TYPE_INT:
      return val, nil
    case PY_TYPE_STRING:
      new_v, err := strconv.ParseInt(val.Data.(string), 10, 10)
      if err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      }
      return VariableType{PY_TYPE_INT, new_v}, nil
    }
    return val, nil
  }
  self.Filters["bool"] = func(val VariableType, args...interface{}) (VariableType, error) {
    b_val, err := val.AsBool()
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, err
    } else {
      return VariableType{PY_TYPE_BOOL, b_val}, nil
    }
  }
}
func (self *Context) LoadDefaultTests() {
  self.Tests["defined"] = func(val VariableType, args...interface{}) (VariableType, error) {
    // FIXME: the jinja2 builtin accepts a value as an arg,
    //        which is returned if the test evaluates to true
    if val.Type == PY_TYPE_UNDEFINED {
      return VariableType{PY_TYPE_BOOL, false}, nil
    } else {
      return VariableType{PY_TYPE_BOOL, true}, nil
    }
  }
}
func NewContext(vars map[string]VariableType) *Context {
  c := new(Context)
  if vars != nil {
    c.Variables = vars
  } else {
    c.Variables = make(map[string]VariableType)
  }
  c.Filters = make(map[string]func(VariableType, ...interface{})(VariableType, error))
  c.Tests = make(map[string]func(VariableType, ...interface{})(VariableType, error))
  c.LoadDefaultFilters()
  c.LoadDefaultTests()
  return c
}

//-------------------------------------------------------------------------------------------------
