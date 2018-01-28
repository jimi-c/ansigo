package main

import(
  "ansible/executor"
  "ansible/playbook"
  action_base "ansible/plugins/action"
)

type ActionPlugin struct {
  action_base.ActionPluginBase
}

func (a *ActionPlugin) Run(task playbook.Task, variables map[string]interface{}) map[string]interface{} {
  executor.CompileModule(task.Action(), variables)
  a.Initialize(task, variables)
  return action_base.ExecuteModule(a, "", nil, "/tmp", nil)
}

var Action ActionPlugin
