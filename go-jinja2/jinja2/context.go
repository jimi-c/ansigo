package jinja2

import (
  "errors"
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
func (self *Context) AddVariables(vars map[string]interface{}) error {
  for k, v := range vars {
    py_v, err := GoVarToPyVar(v)
    if err != nil {
      return err
    }
    self.Variables[k] = py_v
  }
  return nil
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

func InterfaceToPyType(v interface{}) PyType {
  switch v.(type) {
  case string:
    return PY_TYPE_STRING
  case int, int32, int64:
    return PY_TYPE_INT
  case bool:
    return PY_TYPE_BOOL
  case float32, float64:
    return PY_TYPE_FLOAT
  case []interface{}:
    return PY_TYPE_LIST
  case map[interface{}]interface{}:
    return PY_TYPE_DICT
  }
  return PY_TYPE_UNDEFINED
}

func GoVarToPyVar(v interface{}) (VariableType, error) {
  pytype := InterfaceToPyType(v)
  if pytype == PY_TYPE_UNDEFINED {
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("Uknown variable type being added to the context")
  }
  switch pytype {
  case PY_TYPE_STRING:
    v = v.(string)
    return VariableType{PY_TYPE_STRING, v}, nil
  case PY_TYPE_INT:
    v = int64(v.(int))
    return VariableType{PY_TYPE_INT, v}, nil
  case PY_TYPE_BOOL:
    v = v.(bool)
    return VariableType{PY_TYPE_BOOL, v}, nil
  case PY_TYPE_FLOAT:
    v = float64(v.(float64))
    return VariableType{PY_TYPE_FLOAT, v}, nil
  case PY_TYPE_LIST:
    tmp := v.([]interface{})
    res := make([]VariableType, len(tmp))
    for idx, item := range tmp {
      item_res, err := GoVarToPyVar(item)
      if err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      }
      res[idx] = item_res
    }
    return VariableType{PY_TYPE_LIST, res}, nil
  case PY_TYPE_DICT:
    tmp := v.(map[interface{}]interface{})
    res := make(map[VariableType]VariableType)
    for k, v := range tmp {
      k_res, k_err := GoVarToPyVar(k)
      if k_err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, k_err
      }
      v_res, v_err := GoVarToPyVar(v)
      if v_err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, v_err
      }
      res[k_res] = v_res
    }
    return VariableType{PY_TYPE_DICT, res}, nil
  }
  return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("Unknown or unsupported pytype being added to context.")
}
//-------------------------------------------------------------------------------------------------
