package cmd

import (
	"log"
	"strings"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/config"
)

func init() {
	conf, err := config.Open(config.ConfigFile)
	if err != nil {
		conf = config.New()
	}

	cmd := cli.Command{
		Name:   "init",
		Usage:  "Initialize host options",
		Action: initCfg,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "log-dir",
				Value: conf.ArgFetch("log-dir", "/var/log/pebblescape"),
				Usage: "directory to store job logs",
			},
			cli.StringFlag{
				Name:  "state",
				Value: conf.ArgFetch("state", "/var/lib/pebblescape/host-state.bolt"),
				Usage: "path to state file",
			},
			cli.StringFlag{
				Name:  "repo-path",
				Value: conf.ArgFetch("repo-path", "/tmp/repos"),
				Usage: "path for Git repos",
			},
		},
	}

	RegisterCommand(cmd)
}

func initCfg(c *cli.Context) {
	cfg := config.New()

	argKeys := []string{
		"state",
		"log-dir",
		"repo-path",
	}

	cfg.Args = make(map[string]string)

	for _, k := range argKeys {
		if val := c.String(k); val != "" {
			cfg.Args[k] = val
		}
	}

	envKeys := []string{}

	for _, k := range envKeys {
		if val := c.String(k); val != "" {
			envkey := "pebbles_" + strings.Replace(k, "-", "_", -1)
			cfg.Env[strings.ToUpper(envkey)] = val
		}
	}

	err := cfg.WriteTo(config.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}
}
