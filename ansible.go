package main

import (
  "fmt"
  "os"
  "./ansible/executor"
)

func main() {
  if len(os.Args) < 2 {
    fmt.Println("You must specify one or more playbooks to run")
    os.Exit(1)
  }

  pbe := executor.NewPlaybookExecutor(os.Args[1:])
  result := pbe.Run()
  os.Exit(result)
}
