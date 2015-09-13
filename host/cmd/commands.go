package cmd

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/client"
)

var registeredCommands []cli.Command

func RegisterCommand(cmd cli.Command) {
	registeredCommands = append(registeredCommands, cmd)
}

func RegisteredCommands() []cli.Command {
	return registeredCommands
}

func setClientPort(c *cli.Context) error {
	client.SetHost("http://" + c.GlobalString("host") + ":" + string(c.GlobalString("port")))
	return nil
}
