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

func RegisterCommand(cmd cli.Command) {
	registeredCommands = append(registeredCommands, cmd)
}

func RegisteredCommands() []cli.Command {
	return registeredCommands
}

func getApi() *api.Api {
	conf, err := config.Open(config.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	hostApi := api.New(client, conf, log.New(os.Stderr, "", log.Flags()))
	if err := hostApi.ConnectDb(); err != nil {
		log.Fatal(err)
	}

	return hostApi
}
