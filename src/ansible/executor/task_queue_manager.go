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

type WorkerSlot struct {
  res <-chan TaskResult
}

type WorkerJob struct {
  Host inventory.Host
  Task playbook.Task
  PlayContext playbook.PlayContext
}

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
  // go routine stuff
  workers []WorkerSlot
  work_queue chan WorkerJob
  result_queue chan TaskResult
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

  pending_tasks := 0
  for s, t := iterator.GetNextTaskForHost(host, false); s.RunState != ITERATING_COMPLETE; s, t = iterator.GetNextTaskForHost(host, false) {
    if t.Action() == "meta" {
      fmt.Println("META TASK:", t)
    } else {
      tqm.QueueTask(host, *t, *play_context)
      pending_tasks += 1
      for pending_tasks > 0 {
        res := <-tqm.result_queue
        fmt.Println(res)
        pending_tasks -= 1
      }
    }
  }
  fmt.Println("TASK ITERATION COMPLETE FOR HOST: ", host)
  return TQM_RUN_OK
}

func (tqm *TaskQueueManager) QueueTask(host inventory.Host, task playbook.Task, play_context playbook.PlayContext) {
  job := WorkerJob{host, task, play_context}
  tqm.work_queue <- job
  fmt.Println("- queued task")
}

func NewTaskQueueManager(inventory *inventory.InventoryManager, run_additional_callbacks bool) *TaskQueueManager {
  tqm := new(TaskQueueManager)
  tqm.Inventory = inventory
  tqm.Terminated = false
  tqm.StartAtDone = false
  tqm.callbacks_loaded = false
  tqm.run_additional_callbacks = run_additional_callbacks
  tqm.work_queue = make(chan WorkerJob, 5)
  tqm.result_queue = make(chan TaskResult, 5)
  tqm.workers = make([]WorkerSlot, 5)
  for i := 0; i < 5; i++ {
    res_chan := make(chan TaskResult)
    tqm.workers[i] = WorkerSlot{res_chan}
    // fan-out to the workers
    go func() {
      for n := range tqm.work_queue {
        te := NewTaskExecutor(n.Host, n.Task, n.PlayContext)
        res_chan <- te.Run()
      }
    }()
    // fan-in the workers results to a single queue
    go func() {
      for n := range res_chan {
        tqm.result_queue <- n
      }
    }()
  }
  return tqm
}
