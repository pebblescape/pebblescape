package cmd

import (
	"log"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/config"
)

func Init(c *cli.Context) {
	cfg := config.New()

	keys := []string{
		"external-ip",
		"state",
		"log-dir",
		"git-port",
		"repo-path",
		"git-skip-auth",
		"git-keys",
	}

	for _, k := range keys {
		if val := c.String(k); val != "" {
			cfg.Args = append(cfg.Args, k, val)
		}
	}

	err := cfg.WriteTo(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}
}
