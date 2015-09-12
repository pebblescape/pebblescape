package cmd

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
)

var registeredCommands []cli.Command

func RegisterCommand(cmd cli.Command) {
	registeredCommands = append(registeredCommands, cmd)
}

func RegisteredCommands() []cli.Command {
	return registeredCommands
}
