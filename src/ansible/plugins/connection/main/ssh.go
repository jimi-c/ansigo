package main

import (
  "fmt"
  "bytes"
  "io"
  "os/exec"
  //"strings"
  connection_base "ansible/plugins/connection"
)

type ConnectionPlugin struct {
  connection_base.ConnectionPluginBase
  ControlPath string
  ControlPathDir string
}

func (c *ConnectionPlugin) Connect() {
  c.Connected = true
}

func (c *ConnectionPlugin) Close() {
  c.Connected = false
}

func (c *ConnectionPlugin) Execute(cmd []string, in_data string) (int, string, string) {
  ssh_executable := c.PlayContext.SSH_executable()

  // -tt can cause various issues in some environments so allow the user
  // to disable it as a troubleshooting method.
  // FIXME: enable
  //use_tty = self.get_option('use_tty')
  use_tty := false
  // FIXME: sudoable as an option to Execute()
  var args []string
  //if in_data == "" && sudoable && use_tty:
  if in_data == "" && use_tty {
    args = []string{ssh_executable, "-tt", c.Host.Name}
  } else {
    args = []string{ssh_executable, c.Host.Name}
  }
  args = append(args, cmd...)
  the_cmd := BuildCommand(args[0], args[1:])
  fmt.Println("SSH COMMAND:", the_cmd)
  command := exec.Command(the_cmd[0], the_cmd[1:]...)
  stdin, err := command.StdinPipe()
  if err != nil {
    // FIXME: error handling
  }

  var stdout, stderr bytes.Buffer
  command.Stdout = &stdout
  command.Stderr = &stderr
  rc := 0
  if err := command.Start(); err != nil {
    // FIXME: error handling
    rc = 1
	} else {
    if in_data != "" {
      io.WriteString(stdin, in_data)
    }
    stdin.Close()
    if err := command.Wait(); err != nil {
      rc = 1
    }
  }
  return rc, stdout.String(), stderr.String()
}

func (c *ConnectionPlugin) PutFile(in_path string, out_path string) {
}

func (c *ConnectionPlugin) GetFile(in_path string, out_path string) {
}

func BuildCommand(binary string, other_args []string) []string {
  command_parts := make([]string, 0)
  command_parts = append(command_parts, binary)
  command_parts = append(command_parts, []string{"-o", "User=root", "-C", "-o", "ControlMaster=auto", "-o", "ControlPersist=60s", "-o", "ControlPath=/tmp/badbadbad"}...)
  if other_args != nil {
    command_parts = append(command_parts, other_args...)
  }
  return command_parts
}

var Connection ConnectionPlugin
