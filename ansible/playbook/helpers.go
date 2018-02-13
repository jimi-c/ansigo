package playbook

import (
)

var IncludeTasksNames = []string{"include", "include_tasks", "import_tasks"}
var IncludeRolesNames = []string{"include_roles", "import_roles"}

func LoadListOfBlocks(block_data_list []interface{}, play *Play, parent Parent, use_handlers bool) []Block {
  block_list := make([]Block, 0)
  for _, block_data := range block_data_list {
    if data_struct, ok := block_data.(map[interface{}]interface{}); ok {
      new_block := NewBlock(data_struct, play, parent, use_handlers)
      block_list = append(block_list, *new_block)
    }
  }
  return block_list
}

func LoadListOfTasks(task_data_list []interface{}, play *Play, parent Parent, use_handlers bool) []interface{} {
  task_list := make([]interface{}, 0)
  for _, task_data := range task_data_list {
    // FIXME handle errors here
    task_data, _ := task_data.(map[interface{}]interface{})
    _, contains_block := task_data["block"]
    _, contains_rescue := task_data["rescue"]
    _, contains_always := task_data["always"]
    if (contains_block || contains_rescue || contains_always) {
      new_block := NewBlock(task_data, play, parent, use_handlers)
      task_list = append(task_list, *new_block)
    } else {
      // check to see if the data map contains one of the speecial
      // include statement strings. If so, we handle it differently
      var contains_include_tasks bool = false
      for _, include := range IncludeTasksNames {
        _, check := task_data[include]
        contains_include_tasks = contains_include_tasks || check
      }
      var contains_include_roles bool = false
      for _, include := range IncludeRolesNames {
        _, check := task_data[include]
        contains_include_roles = contains_include_roles || check
      }
      if contains_include_tasks {
        // FIXME do include tasks
      } else if contains_include_roles {
        // FIXME do include roles
      } else {
        // No include, so this is a task (or a task in a handlers
        // section of the playbook or role)
        if use_handlers {
          // FIXME do handler stuff
        } else {
          new_task := NewTask(task_data, parent)
          task_list = append(task_list, *new_task)
        }
      }
    }
  }
  return task_list
}
