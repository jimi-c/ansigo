package connection

import (
  "../../inventory"
  "../../playbook"
)

type ConnectionPluginBase struct {
  Host inventory.Host
  Task playbook.Task
  PlayContext playbook.PlayContext
  Connected bool
}

func (c *ConnectionPluginBase) Initialize(host inventory.Host, task playbook.Task, pc playbook.PlayContext) {
  c.Host = host
  c.Task = task
  c.PlayContext = pc
  c.Connected = false
}
