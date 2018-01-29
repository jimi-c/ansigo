package executor

import (
  "fmt"
  "ansible/inventory"
  "ansible/playbook"
)

type PlaybookExecutor struct {
  Inventory *inventory.InventoryManager
  TQM *TaskQueueManager
  Playbooks []string
}

func (pbe *PlaybookExecutor) Load(playbooks []string) {
  pbe.Playbooks = playbooks
  pbe.TQM = NewTaskQueueManager(pbe.Inventory, false)
}

func (pbe *PlaybookExecutor) Run() int {
  result := 0
  break_play := false
  // defer cleanup of TQM here to run after end of function

  // do initial discovery of modules for the builtin modules
  playbook.EnumerateModules("./modules/")

  for _, playbook_path := range pbe.Playbooks {
    pb := playbook.NewPlaybook(playbook_path)
    for play_idx, play := range pb.Entries {
      // set loader basepath
      // clear inventory restriction
      // post-validate the play
      validated_play, ok := playbook.PostValidate(&play).(*playbook.Play)
      if !ok {
        // FIXME: error handling
      }
      _ = play_idx
      // do vars prompting
      // if doing a syntax check, continue
      // if we don't have a tqm save the entry, otherwise we run it
      if pbe.TQM == nil {
        // save it
      } else {
        // run it
        // do accounting to track previously failed/unreachable/etc.
        // self._tqm._unreachable_hosts.update(self._unreachable_hosts)
        previously_failed := 0 //len(self._tqm._failed_hosts)
        previously_unreachable := 0 //len(self._tqm._unreachable_hosts)
        // get serial batches
        serial_batches := pbe.GetSerializedBatches(*validated_play)
        // loop over serial batches
        if len(serial_batches) == 0 {
          fmt.Println("SERIAL BATCHES ARE ZERO")
        } else {
          for _, batch := range serial_batches {
            // set inventory restriction to batch

            // execute TQM Run()
            fmt.Println("running tqm")
            pbe.TQM.Run(validated_play)

            // break the play if the result equals the special return code

            // check the number of failures here, to see if they're above the maximum
            // failure percentage allowed, or if any errors are fatal. If either of those
            // conditions are met, we break out, otherwise we only break out if the entire
            // batch failed
            failed_hosts_count := 0 //len(self._tqm._failed_hosts) + len(self._tqm._unreachable_hosts) - (previously_failed + previously_unreachable)
            if len(batch) == failed_hosts_count {
              break_play = true
              break
            }

            previously_failed += 0 //len(self._tqm._failed_hosts) - previously_failed
            previously_unreachable += 0 //len(self._tqm._unreachable_hosts) - previously_unreachable
            // save the unreachable hosts from this batch
            //self._unreachable_hosts.update(self._tqm._unreachable_hosts)
          }
        }
      }
      if break_play {
        break
      }
    }
    if result != 0 {
      break
    }
  }
  return result
}

func (pbe *PlaybookExecutor) GetSerializedBatches(play playbook.Play) [][]inventory.Host {
  serialized_batches := make([][]inventory.Host, 0)

  all_hosts := make([]inventory.Host, 0)
  // FIXME: stub to create inventory
  for _, h := range play.Hosts() {
    all_hosts = append(all_hosts, *inventory.NewHost(h, nil))
  }
  all_hosts_len := len(all_hosts)

  serial_batch_list := play.Serial()
  if len(serial_batch_list) == 0 {
    serial_batch_list = []int{-1}
  }

  cur_item := 0
  cur_host := 0
  for cur_host < all_hosts_len {
    serial := PctToInt(serial_batch_list[cur_item], all_hosts_len, 1)
    if serial <= 0 {
      serialized_batches = append(serialized_batches, all_hosts[cur_host:])
      break
    } else {
      play_hosts := make([]inventory.Host, 0)
      for x := 0; x < serial; x++ {
        if cur_host > (all_hosts_len - 1) {
          break
        }
        play_hosts = append(play_hosts, all_hosts[cur_host])
        cur_host += 1
      }
      serialized_batches = append(serialized_batches, play_hosts)
    }

    // increment the current batch list item number, and if we've hit
    // the end keep using the last element until we've consumed all of
    // the hosts in the inventory
    cur_item += 1
    if cur_item > len(serial_batch_list) - 1 {
      cur_item = len(serial_batch_list) - 1
    }
  }

  return serialized_batches
}

func NewPlaybookExecutor(playbooks []string) *PlaybookExecutor {
  pbe := new(PlaybookExecutor)
  pbe.Load(playbooks)
  return pbe
}
