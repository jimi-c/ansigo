package executor

import (
  "../inventory"
  "../playbook"
  "../plugins"
)

type TaskResult struct {
  Host inventory.Host
  Task playbook.Task
  Result map[string]interface{}
}

type TaskExecutor struct {
  Host inventory.Host
  Task playbook.Task
  PlayContext playbook.PlayContext
}

func (te *TaskExecutor) Run() TaskResult {
  items := te.GetLoopItems()
  var res map[string]interface{} = nil
  if items != nil {
    // FIXME: all item stuff...
  } else {
    res = te.Execute(nil)
  }

  if _, ok := res["changed"]; !ok {
    res["changed"] = false
  }
  // FIXME: clean result
  // FIXME: close connection

  tr := TaskResult{te.Host, te.Task, res}
  return tr
}

func (te *TaskExecutor) GetLoopItems() []interface{} {
  return nil
}

func (te *TaskExecutor) Execute(vars map[string]interface{}) map[string]interface{} {
  variables := vars
  if variables == nil {
    // FIXME: use variables from struct, but this only matters for per-item stuff anyway
    variables = make(map[string]interface{})
  }

  // FIXME: play context updating and validation
  // apply the given task's information to the connection info,
  // which may override some fields already set by the play or
  // the options specified on the command line
  //self._play_context = self._play_context.set_task_and_variable_override(task=self._task, variables=variables, templar=templar)

  // fields set from the play/task may be based on variables, so we have to
  // do the same kind of post validation step on it here before we use it.
  //self._play_context.post_validate(templar=templar)

  //now that the play context is finalized, if the remote_addr is not set
  // default to using the host's address field as the remote address
  //if not self._play_context.remote_addr:
  //    self._play_context.remote_addr = self._host.address

  // We also add "magic" variables back into the variables dict to make sure
  // a certain subset of variables exist.
  //self._play_context.update_vars(variables)

  // FIXME: update connection/shell plugin options

  if !te.Task.EvaluateConditional() {
    // FIXME: add no_log field in later
    return map[string]interface{} {
      "changed": false,
      "skipped": true,
      "skip_reason": "Conditional result was False",
    }
  }

  // FIXME: implement loop eval and context validation error handling here
  // FIXME: include/include_task/include_role handling here

  connection := te.GetConnection()
  handler := te.GetActionHandler(connection)

  retries := 1
  if te.Task.Until() != nil {
    retries = te.Task.Retries()
    if retries < 0 {
      retries = 1
    } else {
      retries += 1
    }
  }

  delay := te.Task.Delay()
  if delay < 0 {
    delay = 1
  }

  var res map[string]interface{} = nil
  for i := 1; i < retries + 1; i++ {
    res = handler.Run(te.Task, variables)
    if te.Task.Register() != "" {
      // FIXME: clean/wrap res here
      variables[te.Task.Register()] = res
    }
    // FIXME: preserve no_log
    // FIXME: do async handling
    if _, ok := res["failed"]; !ok {
      // FIXME:
      if v, ok := res["rc"]; ok {
        switch v {
        case 0:
          res["failed"] = false
        case "0":
          res["failed"] = false
        default:
          res["failed"] = true
        }
      } else {
        res["failed"] = false
      }
    }
    if _, ok := res["changed"]; !ok {
      res["changed"] = false
    }
    // FIXME: do retry handling
    // FIXME: do changed_when/failed_when handling
  }
  if retries > 1 {
    res["attempts"] = retries + 1
    res["failed"] = true
  }
  // FIXME: wrap ansible_facts in result
  notify := te.Task.Notify()
  if notify != nil {
    res["_ansible_notify"] = notify
  }
  // FIXME: preserve the delegated_vars
  return res
}

func (te *TaskExecutor) GetConnection() plugins.ConnectionInterface {
  conn_name := te.PlayContext.Connection()
  if conn_name == "smart" {
    // FIXME: control persist detection
    conn_name = "ssh"
  }
  conn := plugins.LoadConnectionPlugin(conn_name)
  conn.Initialize(te.Host, te.Task, te.PlayContext)
  return conn
}

func (te *TaskExecutor) GetActionHandler(connection plugins.ConnectionInterface) plugins.ActionInterface {
  var handler plugins.ActionInterface
  if plugins.PluginExists(te.Task.Action(), "action") {
    handler = plugins.LoadActionPlugin(te.Task.Action())
  } else {
    handler = plugins.LoadActionPlugin("normal")
  }
  handler.SetConnection(connection)
  return handler
}

func NewTaskExecutor(host inventory.Host, task playbook.Task, pc playbook.PlayContext) *TaskExecutor {
  te := new(TaskExecutor)
  te.Host = host
  te.Task = task
  te.PlayContext = pc
  return te
}
