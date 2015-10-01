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
	// "github.com/pebblescape/pebblescape/host/types"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
)

func init() {
	conf, err := config.Open(config.ConfigFile)
	if err != nil {
		conf = config.New()
	}

	cmd := cli.Command{
		Name:   "daemon",
		Usage:  "Start host daemon",
		Action: daemon,
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

func daemon(c *cli.Context) {
	var logger *log.Logger

	log.SetFlags(log.Ldate | log.Lmicroseconds)

	port := c.GlobalString("port")
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

	state := state.NewState(stateFile)

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	backend, err := backend.NewDockerBackend(client, state)
	if err != nil {
		log.Fatal(err)
	}

	shutdown.BeforeExit(func() {
		if err := backend.Cleanup(); err != nil {
			log.Fatal(err)
		}
	})

	resurrect, err := state.Restore(backend)
	if err != nil {
		log.Fatal(err)
	}

	resurrect()

	// job := &host.Job{
	// 	Config: &docker.Config{
	// 		Image: "ubuntu",
	// 		Cmd: []string{
	// 			"/bin/bash",
	// 			"-c",
	// 			"ping google.com",
	// 		},
	// 	},
	// }
	// err = state.RunJob(job)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	http.Serve(port, state, gitRepos, logger)
}
