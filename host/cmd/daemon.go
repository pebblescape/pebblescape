package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/backend"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/http"
	"github.com/pebblescape/pebblescape/host/state"
)

func init() {
	conf, err := config.Open(config.ConfigFile)
	if err != nil {
		conf = config.New()
	}

	cmd := cli.Command{
		Name:   "daemon",
		Usage:  "start host daemon",
		Action: Daemon,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "log-dir",
				Value: conf.ArgFetch("log-dir", ""),
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

func Daemon(c *cli.Context) {
	var logger *log.Logger

	logDir := c.String("log-dir")
	stateFile := c.String("state")
	gitRepos := c.String("repo-path")

	if logDir != "" {
		logFile := filepath.Join(logDir, "host.log")
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			log.Fatalf("could not not mkdir for logs: %s", err)
		}

		hostlog, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer hostlog.Close()

		log.Printf("Logging to %s\n", logFile)
		log.SetOutput(io.MultiWriter(hostlog, os.Stdout))
		logger = log.New(io.MultiWriter(hostlog, os.Stderr), "", log.Flags())
	} else {
		logger = log.New(os.Stderr, "", log.Flags())
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	backend, err := backend.NewDockerBackend(client)
	if err != nil {
		log.Fatal(err)
	}

	state := state.NewState(stateFile)
	err = state.Restore(backend)
	if err != nil {
		log.Fatal(err)
	}

	http.Serve(state, gitRepos, logger)
}
