package main

import (
	"log"
	"os"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/cmd"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/pkg/version"
)

func init() {
	log.SetFlags(0)
}

func main() {
	conf, err := config.Open(config.ConfigFile)
	if err != nil {
		conf = config.New()
	}

	for k, v := range conf.Env {
		os.Setenv(k, v)
	}

	app := cli.NewApp()
	app.Name = "pebblescape"
	app.Usage = "manage pebblescape host"
	app.Version = version.String()
	app.Commands = cmd.RegisteredCommands()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Value: "localhost",
			Usage: "Pebblescape host",
		},
		cli.StringFlag{
			Name:  "port, p",
			Value: "4592",
			Usage: "Pebblescape port",
		},
	}
	app.Run(os.Args)
}
