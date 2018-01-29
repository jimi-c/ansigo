package inventory

import ()

type Host struct {
  Name string
  Vars map[string]interface{}
}

func NewHost(name string, vars map[string]interface{}) *Host {
  h := new(Host)
  h.Name = name
  if vars != nil {
    h.Vars = vars
  } else {
    h.Vars = make(map[string]interface{})
  }
  return h
}
