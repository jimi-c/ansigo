package main

import(
  //"ansible/executor"
  "github.com/ansible/ansigo/src/playbook"
  action_base "github.com/ansible/ansigo/src/plugins/action"
)

type ActionPlugin struct {
  action_base.ActionPluginBase
}

func (a *ActionPlugin) Run(task playbook.Task, variables map[string]interface{}) map[string]interface{} {
  a.Initialize(task, variables)
  return map[string]interface{}{"msg": "debugged"}
}

var Action ActionPlugin
