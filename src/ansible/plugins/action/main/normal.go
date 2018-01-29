package main

import(
  //"ansible/executor"
  "ansible/playbook"
  action_base "ansible/plugins/action"
)

type ActionPlugin struct {
  action_base.ActionPluginBase
}

func (a *ActionPlugin) Run(task playbook.Task, variables map[string]interface{}) map[string]interface{} {
  a.Initialize(task, variables)
  return action_base.ExecuteModule(a, "", task.Args(), "/tmp", nil)
}

var Action ActionPlugin
