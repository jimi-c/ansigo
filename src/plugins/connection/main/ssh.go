package main

type ConnectionPlugin struct {

}

func (c *ConnectionPlugin) Connect() {

}

func (c *ConnectionPlugin) Close() {

}

func (c *ConnectionPlugin) Execute() (int, string, string) {
  return 0, "stdout from ssh connection", "stderr from ssh connection"
}

func (c *ConnectionPlugin) PutFile() {

}

func (c *ConnectionPlugin) GetFile() {

}

var Connection ConnectionPlugin
