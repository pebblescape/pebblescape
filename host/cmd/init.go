package cmd

import (
	"log"
	"strings"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/config"
)

func Init(c *cli.Context) {
	cfg := config.New()

	argKeys := []string{
		"external-ip",
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
