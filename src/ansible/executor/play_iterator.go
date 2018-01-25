package executor

import (
  "ansible/inventory"
  "ansible/playbook"
)

// the primary running states for the play iteration
const ITERATING_SETUP = 0
const ITERATING_TASKS = 1
const ITERATING_RESCUE = 2
const ITERATING_ALWAYS = 3
const ITERATING_COMPLETE = 4

// the failure states for the play iteration, which are powers
// of 2 as they may be or'ed together in certain circumstances
const ITERATING_FAILED_NONE = 0
const ITERATING_FAILED_SETUP = 1
const ITERATING_FAILED_TASKS = 2
const ITERATING_FAILED_RESCUE = 4
const ITERATING_FAILED_ALWAYS = 8

type HostState struct {
  Blocks []playbook.Block

  CurBlock int
  CurRegularTask int
  CurRescueTask int
  CurAlwaysTask int
  CurDepChain []interface{}
  RunState int
  FailState int
  PendingSetup bool
  TasksChildState *HostState
  RescueChildState *HostState
  AlwaysChildState *HostState
  DidRescue bool
  DidStartAtTask bool
}

func (hs *HostState) Copy() *HostState{
  new_hs := NewHostState(hs.Blocks)
  new_hs.CurBlock = hs.CurBlock
  new_hs.CurRegularTask = hs.CurRegularTask
  new_hs.CurRescueTask = hs.CurRescueTask
  new_hs.CurAlwaysTask = hs.CurAlwaysTask
  new_hs.RunState = hs.RunState
  new_hs.FailState = hs.FailState
  new_hs.PendingSetup = hs.PendingSetup
  new_hs.DidRescue = hs.DidRescue
  new_hs.DidStartAtTask = hs.DidStartAtTask

  new_hs.CurDepChain = make([]interface{}, len(hs.CurDepChain))
  copy(new_hs.CurDepChain, hs.CurDepChain)

  if hs.TasksChildState != nil {
    new_hs.TasksChildState = hs.TasksChildState.Copy()
  }
  if hs.RescueChildState != nil {
    new_hs.RescueChildState = hs.RescueChildState.Copy()
  }
  if hs.AlwaysChildState != nil {
    new_hs.AlwaysChildState = hs.AlwaysChildState.Copy()
  }
  return new_hs
}

func NewHostState(blocks []playbook.Block) *HostState {
  s := new(HostState)
  s.Blocks = make([]playbook.Block, len(blocks))
  copy(s.Blocks, blocks)
  s.CurBlock = 0
  s.CurRegularTask = 0
  s.CurRescueTask = 0
  s.CurAlwaysTask = 0
  s.CurDepChain = nil
  s.RunState = ITERATING_SETUP
  s.FailState = ITERATING_FAILED_NONE
  s.PendingSetup = false
  s.TasksChildState = nil
  s.RescueChildState = nil
  s.AlwaysChildState = nil
  s.DidRescue = false
  s.DidStartAtTask = false
  return s
}

type PlayIterator struct {
  Play *playbook.Play
  PlayContext *playbook.PlayContext
  Inventory *inventory.InventoryManager
  BatchSize int
  HostStates map[string]*HostState
  var_manager interface{}
  blocks []playbook.Block
}

func NewPlayIterator(
    tqm *TaskQueueManager,
    play *playbook.Play,
    play_context *playbook.PlayContext,
    all_vars map[string]interface{},
  ) *PlayIterator {

  iterator := new(PlayIterator)
  iterator.Play = play
  iterator.Inventory = tqm.Inventory
  iterator.var_manager = tqm.VarManager
  iterator.HostStates = make(map[string]*HostState)

  dummy_block_data := make(map[interface{}]interface{})
  setup_block := playbook.NewBlock(dummy_block_data, play, nil, false)
  dummy_task_data := make(map[interface{}]interface{})
  dummy_task_data["setup"] = make(map[interface{}]interface{})
  // FIXME: extra setup args here
  setup_task := playbook.NewTask(dummy_task_data, setup_block)
  setup_block.Attr_block = append(setup_block.Attr_block, setup_task)

  iterator.blocks = append(iterator.blocks, *setup_block)
  // FIXME: filter blocks based on tags here
  iterator.blocks = append(iterator.blocks, play.Compile()...)

  //start_at_matched := false
  batch := tqm.Inventory.GetHosts()
  iterator.BatchSize = len(batch)
  for _, host := range batch {
    iterator.HostStates[host.Name] = NewHostState(iterator.blocks)
    // FIXME: do start_at_task stuff here
  }
  return iterator
}

func (it *PlayIterator) GetHostState(host inventory.Host) *HostState {
  state, ok := it.HostStates[host.Name]
  if !ok {
    state = NewHostState(make([]playbook.Block, 0))
    it.HostStates[host.Name] = state
  }
  return state.Copy()
}

func (it *PlayIterator) GetNextTaskForHost(host inventory.Host, peek bool) (*HostState, *playbook.Task) {
  s := it.GetHostState(host)
  var t *playbook.Task = nil
  if s.RunState == ITERATING_COMPLETE {
    return s, nil
  }
  s, t = it.GetNextTaskFromState(s, host, peek, false)
  if !peek {
    it.HostStates[host.Name] = s
  }
  return s, t
}

func (it *PlayIterator) GetNextTaskFromState(state *HostState, host inventory.Host, peek bool, in_child bool) (*HostState, *playbook.Task) {
  var task *playbook.Task = nil

  for {
    if state.CurBlock >= len(state.Blocks) {
      state.RunState = ITERATING_COMPLETE
      return state, nil
    }

    cur_block := state.Blocks[state.CurBlock]
    if state.RunState == ITERATING_SETUP {
      if !state.PendingSetup {
        state.PendingSetup = true

        // Gather facts if the default is 'smart' and we have not yet
        // done it for this host; or if 'explicit' and the play sets
        // gather_facts to True; or if 'implicit' and the play does
        // NOT explicitly set gather_facts to False.

        gathering := "smart" // FIXME: C.DEFAULT_GATHERING
        implied := it.Play.Attr_gather_facts == nil || it.Play.GatherFacts()

        // FIXME: below
        //if (gathering == "implicit" and implied) ||
        //   (gathering == "explicit" and it.Play.GatherFacts()) ||
        //   (gathering == "smart" && implied && !(self._variable_manager._fact_cache.get(host.name, {}).get('module_setup', False))) {
        if (gathering == "implicit" && implied) ||
           (gathering == "explicit" && it.Play.GatherFacts()) ||
           (gathering == "smart" && implied) {
          // The setup block is always self._blocks[0], as we inject it
          // during the play compilation in __init__ above.
          setup_block := state.Blocks[0]
          // FIXME: below
          // if setup_block.has_tasks() && len(setup_block.block) > 0 {
          if len(setup_block.Attr_block) > 0 {
            the_task := setup_block.Attr_block[0].(playbook.Task)
            task = &the_task
          }
        }
      } else {
        // This is the second trip through ITERATING_SETUP, so we clear
        // the flag and move onto the next block in the list while setting
        // the run state to ITERATING_TASKS
        state.PendingSetup = false
        state.RunState = ITERATING_TASKS
        if !state.DidStartAtTask {
          state.CurBlock += 1
          state.CurRegularTask = 0
          state.CurRescueTask = 0
          state.CurAlwaysTask = 0
          state.TasksChildState = nil
          state.RescueChildState = nil
          state.AlwaysChildState = nil
        }
      }
    } else if state.RunState == ITERATING_TASKS {
      // clear the pending setup flag, since we're past that and it didn't fail
      if state.PendingSetup {
        state.PendingSetup = false
      }
      // First, we check for a child task state that is not failed, and if we
      // have one recurse into it for the next task. If we're done with the child
      // state, we clear it and drop back to getting the next task from the list.
      if state.TasksChildState != nil {
        state.TasksChildState, task = it.GetNextTaskFromState(state.TasksChildState, host, peek, true)
        if it.CheckFailedState(state.TasksChildState) {
          // failed child state, so clear it and move into the rescue portion
          state.TasksChildState = nil
          it.SetFailedState(state)
        } else {
          // get the next task recursively
          if task == nil || state.TasksChildState.RunState == ITERATING_COMPLETE {
            // we're done with the child state, so clear it and continue
            // back to the top of the loop to get the next task
            state.TasksChildState = nil
            continue
          }
        }
      } else {
        // First here, we check to see if we've failed anywhere down the chain
        // of states we have, and if so we move onto the rescue portion. Otherwise,
        // we check to see if we've moved past the end of the list of tasks. If so,
        // we move into the always portion of the block, otherwise we get the next
        // task from the list.
        if it.CheckFailedState(state) {
          state.RunState = ITERATING_RESCUE
        } else if state.CurRegularTask >= len(cur_block.Attr_block) {
          state.RunState = ITERATING_ALWAYS
        } else {
          thing := cur_block.Attr_block[state.CurRegularTask]
          // if the current task is actually a child block, create a child
          // state for us to recurse into on the next pass
          child_block, is_block := thing.(playbook.Block)
          if is_block || state.TasksChildState != nil {
            state.TasksChildState = NewHostState([]playbook.Block{child_block})
            state.TasksChildState.RunState = ITERATING_TASKS
            // since we've created the child state, clear the task
            // so we can pick up the child state on the next pass
            task = nil
          } else {
            the_task := thing.(playbook.Task)
            task = &the_task
          }
          state.CurRegularTask += 1
        }
      }
    } else if state.RunState == ITERATING_RESCUE {
      // The process here is identical to ITERATING_TASKS, except
      // instead we move into the always portion of the block.
      // We also clear this host from the list of hosts removed from
      // the play due to a failure.
      if _, ok := it.Play.RemovedHosts[host.Name]; ok && !peek {
        delete(it.Play.RemovedHosts, host.Name)
      }

      if state.RescueChildState != nil {
        state.RescueChildState, task = it.GetNextTaskFromState(state.RescueChildState, host, peek, true)
        if it.CheckFailedState(state.RescueChildState) {
          state.RescueChildState = nil
          it.SetFailedState(state)
        } else {
          if task == nil || state.RescueChildState.RunState == ITERATING_COMPLETE {
            state.RescueChildState = nil
            continue
          }
        }
      } else {
        if state.FailState & ITERATING_FAILED_RESCUE == ITERATING_FAILED_RESCUE {
          state.RunState = ITERATING_ALWAYS
        } else if state.CurRescueTask >= len(cur_block.Attr_rescue) {
          if len(cur_block.Attr_rescue) > 0 {
            state.FailState = ITERATING_FAILED_NONE
          }
          state.RunState = ITERATING_ALWAYS
          state.DidRescue = true
        } else {
          thing := cur_block.Attr_rescue[state.CurRescueTask]
          // if the current task is actually a child block, create a child
          // state for us to recurse into on the next pass
          child_block, is_block := thing.(playbook.Block)
          if is_block || state.RescueChildState != nil {
            state.RescueChildState = NewHostState([]playbook.Block{child_block})
            state.RescueChildState.RunState = ITERATING_TASKS
            task = nil
          } else {
            the_task := thing.(playbook.Task)
            task = &the_task
          }
          state.CurRescueTask += 1
        }
      }
    } else if state.RunState == ITERATING_ALWAYS {
      // And again, the process here is identical to ITERATING_TASKS, except
      // instead we either move onto the next block in the list, or we set the
      // run state to ITERATING_COMPLETE in the event of any errors, or when we
      // have hit the end of the list of blocks.
      if state.AlwaysChildState != nil {
        state.AlwaysChildState, task = it.GetNextTaskFromState(state.AlwaysChildState, host, peek, true)
        if it.CheckFailedState(state.AlwaysChildState) {
          state.AlwaysChildState = nil
          it.SetFailedState(state)
        } else {
          if task == nil || state.AlwaysChildState.RunState == ITERATING_COMPLETE {
            state.AlwaysChildState = nil
            continue
          }
        }
      } else {
        if state.CurAlwaysTask >= len(cur_block.Attr_always) {
          if state.FailState != ITERATING_FAILED_NONE {
            state.RunState = ITERATING_COMPLETE
          } else {
            state.CurBlock += 1
            state.CurRegularTask = 0
            state.CurRescueTask = 0
            state.CurAlwaysTask = 0
            state.RunState = ITERATING_TASKS
            state.TasksChildState = nil
            state.RescueChildState = nil
            state.AlwaysChildState = nil
            state.DidRescue = false
            // FIXME: implement when roles are done
            // we're advancing blocks, so if this was an end-of-role block we
            // mark the current role complete
            //if block._eor and host.name in block._role._had_task_run and not in_child and not peek:
            //    block._role._completed[host.name] = True
          }
        } else {
          thing := cur_block.Attr_always[state.CurAlwaysTask]
          child_block, is_block := thing.(playbook.Block)
          if is_block || state.AlwaysChildState != nil {
            state.AlwaysChildState = NewHostState([]playbook.Block{child_block})
            state.AlwaysChildState.RunState = ITERATING_TASKS
            task = nil
          } else {
            the_task := thing.(playbook.Task)
            task = &the_task
          }
          state.CurAlwaysTask += 1
        }
      }
    } else if state.RunState == ITERATING_COMPLETE {
      return state, nil
    }

    if task != nil {
      break
    }
  }
  return state, task
}

func (it *PlayIterator) CheckFailedState(state *HostState) bool {
  return false
}

func (it *PlayIterator) SetFailedState(state *HostState) {

}
