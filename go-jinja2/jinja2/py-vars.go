package jinja2

import (
  "errors"
)

type PyType int
const (
  PY_TYPE_UNDEFINED PyType = 0
  PY_TYPE_NONE      PyType = 1
  PY_TYPE_STRING    PyType = 2
  PY_TYPE_INT       PyType = 3
  PY_TYPE_FLOAT     PyType = 4
  PY_TYPE_BOOL      PyType = 5
  PY_TYPE_LIST      PyType = 6
  PY_TYPE_TUPLE     PyType = 7
  PY_TYPE_DICT      PyType = 8
  PY_TYPE_IDENT     PyType = 10
)

func PyTypeToString(v PyType) string {
  switch v {
  case PY_TYPE_UNDEFINED:
    return "UNDEFINED"
  case PY_TYPE_NONE:
    return "None"
  case PY_TYPE_STRING:
    return "string"
  case PY_TYPE_INT:
    return "int"
  case PY_TYPE_FLOAT:
    return "float"
  case PY_TYPE_BOOL:
    return "bool"
  case PY_TYPE_LIST:
    return "list"
  case PY_TYPE_TUPLE:
    return "tuple"
  case PY_TYPE_DICT:
    return "dict"
  }
  return ""
}

type VariableType struct {
  Type PyType
  Data interface{}
}
func (self *VariableType) AsBool() (bool, error) {
  switch res := self.Type; res {
  case PY_TYPE_NONE:
    return false, nil
  case PY_TYPE_BOOL:
    return self.Data.(bool), nil
  case PY_TYPE_INT:
    if v, ok := self.Data.(int64); ok {
      if v == 0 { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(float64); ok {
      if v == 0.0 { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  case PY_TYPE_STRING:
    if v, ok := self.Data.(string); ok {
      if v == "" { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  }
  return false, errors.New("unknown varaible type, cannot convert it to a boolean value")
}
func (self *VariableType) AsInt() (int64, error) {
  switch res := self.Type; res {
  case PY_TYPE_INT:
    if v, ok := self.Data.(int64); ok {
      return v, nil
    } else {
      return int64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(float64); ok {
      return int64(v), nil
    } else {
      return int64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_BOOL:
    if v, ok := self.Data.(bool); ok {
      if v {
        return int64(1), nil
      } else {
        return int64(0), nil
      }
    } else {
      return int64(0), errors.New("could not convert variable to a string")
    }
  default:
    return int64(0), errors.New("could not convert variable to a string")
  }
}
func (self *VariableType) AsFloat() (float64, error) {
  switch res := self.Type; res {
  case PY_TYPE_INT:
    if v, ok := self.Data.(int64); ok {
      return float64(v), nil
    } else {
      return float64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(float64); ok {
      return v, nil
    } else {
      return float64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_BOOL:
    if v, ok := self.Data.(bool); ok {
      if v {
        return float64(1.0), nil
      } else {
        return float64(0.0), nil
      }
    } else {
      return float64(0.0), errors.New("could not convert variable to a string")
    }
  default:
    return float64(0.0), errors.New("could not convert variable to a string")
  }
}
func (self *VariableType) AsString() (string, error) {
  switch res := self.Type; res {
  case PY_TYPE_STRING:
    if v, ok := self.Data.(string); ok {
      return v, nil
    } else {
      return "", errors.New("could not convert variable to a string")
    }
  default:
    return "", errors.New("could not convert variable to a string")
  }
}
