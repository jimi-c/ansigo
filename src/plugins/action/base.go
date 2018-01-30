package action

import(
  "ansible/executor"
  "github.com/ansible/ansigo/src/playbook"
  "github.com/ansible/ansigo/src/plugins"
)

type ActionPluginBase struct {
  connection plugins.ConnectionInterface
  task playbook.Task
  task_args map[string]interface{}
}

func (a *ActionPluginBase) Initialize(
    task playbook.Task,
    args map[string]interface{},
  ) {
  a.SetTask(task)
  a.SetTaskArgs(args)
}

func (a *ActionPluginBase) Connection() plugins.ConnectionInterface { return a.connection }
func (a *ActionPluginBase) SetConnection(conn plugins.ConnectionInterface) {
  a.connection = conn
}

func (a *ActionPluginBase) Task() playbook.Task { return a.task }
func (a *ActionPluginBase) SetTask(task playbook.Task) { a.task = task }

func (a *ActionPluginBase) TaskArgs() map[string]interface{} { return a.task_args }
func (a *ActionPluginBase) SetTaskArgs(args map[string]interface{}) { a.task_args = args }

func ExecuteModule(
    a plugins.ActionInterface,
    module_name string,
    module_args map[string]interface{},
    tmp string,
    task_vars  map[string]interface{},
  ) map[string]interface{} {

  if module_name == "" {
    task := a.Task()
    module_name = task.Action()
  }
  if module_args == nil { module_args = a.TaskArgs() }
  if task_vars == nil { task_vars = make(map[string]interface{}) }

  module_data := executor.CompileModule(module_name, module_args)
  return LowLevelExecuteCommand(a, []string{"/usr/bin/python"}, module_data)
}

// FIXME: all options
func LowLevelExecuteCommand(a plugins.ActionInterface, cmd []string, in_data string) map[string]interface{} {
  rc, stdout, stderr := a.Connection().Execute(cmd, in_data)
  return map[string]interface{} {
    "rc": rc,
    "stdout": stdout,
    "stderr": stderr,
  }
}
