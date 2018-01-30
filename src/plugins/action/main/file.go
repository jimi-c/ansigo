package main

type ActionPlugin struct {
}

func (a *ActionPlugin) Run(variables map[string]interface{}) map[string]interface{} {
  return map[string]interface{} {
    "msg": "Hello from the action plugin!",
  }
}

var Action ActionPlugin
