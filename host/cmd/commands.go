package cmd

import (
	"log"
	"os"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/api"
	"github.com/pebblescape/pebblescape/host/config"
)

var registeredCommands []cli.Command
var host *api.Api

func RegisterCommand(cmd cli.Command) {
	registeredCommands = append(registeredCommands, cmd)
}

func RegisteredCommands() []cli.Command {
	return registeredCommands
}

func setApi(c *cli.Context) error {
	conf, err := config.Open(config.ConfigFile)
	if err != nil {
		log.Fatal("Host is not running or invalid config file")
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	host = api.New(client, conf, log.New(os.Stderr, "", log.Flags()))
	if err := host.ConnectDb(); err != nil {
		return err
	}

	return nil
}

func fatal(e interface{}) {
	host.Logger.Fatal(e)
}
