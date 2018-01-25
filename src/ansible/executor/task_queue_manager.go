package executor

import (
  "fmt"
  "ansible/inventory"
  "ansible/playbook"
)

const TQM_RUN_OK = 0
const TQM_RUN_ERROR = 1
const TQM_RUN_FAILED_HOSTS = 2
const TQM_RUN_UNREACHABLE_HOSTS = 4
const TQM_RUN_FAILED_BREAK_PLAY = 8
const TQM_RUN_UNKNOWN_ERROR = 255

type CallbackArgs struct {
  what interface{}
}

type TaskQueueManager struct {
  Inventory *inventory.InventoryManager
  VarManager interface{} // FIXME
  Options interface{} // FIXME
  Stats interface{} // FIXME
  Passwords []string
  Terminated bool
  StartAtDone bool
  // handler mappings for notifications
  NotifiedHandlers map[string][]inventory.Host
  ListeningHandlers map[string][]inventory.Host
  // maps to track failed and unreachable hosts
  FailedHosts map[string]bool
  UnreachableHosts map[string]bool
  // private stuff
  std_out_callback interface{} // FIXME
  callbacks_loaded bool
  callback_plugins []interface{} // FIXME
  run_additional_callbacks bool
}

func (tqm *TaskQueueManager) LoadCallbacks() {
  tqm.callbacks_loaded = true
}

func (tqm *TaskQueueManager) SendCallback(name string, args ...CallbackArgs) {

}

func (tqm *TaskQueueManager) Run(play *playbook.Play) int {
  if !tqm.callbacks_loaded {
    tqm.LoadCallbacks()
  }

  // create HostVars

  // create the play context object and assign it to any callback
  // plugins we've loaded that may need it
  play_context := playbook.NewPlayContext(play, tqm.Options, tqm.Passwords)
  //for callback_plugin in self._callback_plugins:
  //  if hasattr(callback_plugin, 'set_play_context'):
  //    callback_plugin.set_play_context(play_context)
  tqm.SendCallback("v2_playbook_on_play_start", CallbackArgs{play})
  // initialize the shared dictionary containing the notified handlers
  //self._initialize_notified_handlers(new_play)
  iterator := NewPlayIterator(tqm, play, play_context, make(map[string]interface{}))
  host := tqm.Inventory.GetHosts()[0]

  for s, t := iterator.GetNextTaskForHost(host, false); s.RunState != ITERATING_COMPLETE; s, t = iterator.GetNextTaskForHost(host, false) {
    fmt.Println("- TASK:", t)
  }
  fmt.Println("TASK ITERATION COMPLETE FOR HOST: ", host)
  return TQM_RUN_OK
}

func NewTaskQueueManager(inventory *inventory.InventoryManager, run_additional_callbacks bool) *TaskQueueManager {
  tqm := new(TaskQueueManager)
  tqm.Inventory = inventory
  tqm.Terminated = false
  tqm.StartAtDone = false
  tqm.callbacks_loaded = false
  tqm.run_additional_callbacks = run_additional_callbacks
  return tqm
}
