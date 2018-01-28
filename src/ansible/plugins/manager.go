package plugins

import (
  "plugin"
  "strings"
  "ansible/playbook"
)

func LoadPlugin(name string, class string) interface{} {
  mod_name := "./plugins/" + class + "/" + name + ".so"
  mod, err := plugin.Open(mod_name)
  if err != nil {
    panic(err)
  }

  the_plugin, err := mod.Lookup(strings.Title(class))
  if err != nil {
    panic(err)
  }

  return the_plugin
}

type ActionInterface interface {
  Run(playbook.Task, map[string]interface{}) map[string]interface{}
  Connection() ConnectionInterface
  SetConnection(ConnectionInterface)
  Task() playbook.Task
  SetTask(playbook.Task)
  TaskArgs() map[string]interface{}
  SetTaskArgs(map[string]interface{})
}

func LoadActionPlugin(name string) ActionInterface {
  return LoadPlugin(name, "action").(ActionInterface)
}

type ConnectionInterface interface {
  Connect()
  Close()
  Execute([]string, string) (int, string, string)
  PutFile(string, string)
  GetFile(string, string)
}

func LoadConnectionPlugin(name string) ConnectionInterface {
  return LoadPlugin(name, "connection").(ConnectionInterface)
}
