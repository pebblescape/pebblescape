package main

import (
	"log"
	"os"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/cmd"
	"github.com/pebblescape/pebblescape/pkg/version"
)

func init() {
	log.SetFlags(0)
}

func main() {
	app := cli.NewApp()
	app.Name = "pebblescape"
	app.Usage = "manage pebblescape host"
	app.Version = version.String()
	app.Commands = cmd.RegisteredCommands()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "/etc/pebblescape/host.json",
			Usage: "host configuration path",
		},
	}
	app.Run(os.Args)
}
