package main

import(
  "bytes"
  "os"
  "os/exec"
  "strings"
  connection_base "ansible/plugins/connection"
)

type ConnectionPlugin struct {
  connection_base.ConnectionPluginBase
}

func (c *ConnectionPlugin) Connect() {
  c.Connected = true
}

func (c *ConnectionPlugin) Close() {
  c.Connected = false
}

func (c *ConnectionPlugin) Execute(cmd []string, in_data string) (int, string, string) {
  executable := cmd[0]
  args := make([]string, 0)
  if len(cmd) > 1 {
    args = append(args, cmd[1:]...)
  }
  command := exec.Command(executable, args...)
  if in_data != "" {
    command.Stdin = strings.NewReader(in_data)
  }
  var stdout, stderr bytes.Buffer
  command.Stdout = &stdout
  command.Stderr = &stderr
  rc := 0
  err := command.Run()
	if err != nil {
    // FIXME: error handling
    rc = 1
	}
  return rc, stdout.String(), stderr.String()
}

func (c *ConnectionPlugin) PutFile(in_path string, out_path string) {
  os.Rename(in_path, out_path)
}

func (c *ConnectionPlugin) GetFile(in_path string, out_path string) {
  c.PutFile(in_path, out_path)
}

// All connection plugins must define this line, it is the entry point
var Connection ConnectionPlugin
