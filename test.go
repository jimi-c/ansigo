package main

import (
  "fmt"
  "io/ioutil"
  "github.com/smallfish/simpleyaml"
  "os"
  "ansible/playbook"
)

func main() {
  if len(os.Args) < 2 {
    fmt.Println("You must specify one or more playbooks to run")
    os.Exit(1)
  }
  for i := 1; i < len(os.Args); i++ {
    yamlFile, err := ioutil.ReadFile(os.Args[i])

    playbook_data, err := simpleyaml.NewYaml(yamlFile)
    if err != nil {
      fmt.Println("Error parsing YAML:", err)
      os.Exit(1)
    }

    plays, err := playbook_data.Array()
    if err != nil {
      fmt.Println("Error: you must specify your plays as a list")
      os.Exit(1)
    }
    for play_idx := 0; play_idx < len(plays); play_idx++ {
      play_data, _ := playbook_data.GetIndex(play_idx).Map()
      p := playbook.NewPlay(play_data)
      fmt.Printf("%#f\n", p)
    }
  }
}
