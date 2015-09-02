package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/cmd"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
	"github.com/pebblescape/pebblescape/pkg/version"
)

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)
}

func main() {
	defer shutdown.Exit()

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

	stateFlag := cli.StringFlag{
		Name:  "state",
		Value: conf.ArgFetch("state", "/var/lib/pebblescape/host-state.bolt"),
		Usage: "path to state file",
	}
	repoFlag := cli.StringFlag{
		Name:  "repo-path",
		Value: conf.ArgFetch("repo-path", "/tmp/repos"),
		Usage: "path for Git repo caches",
	}

	app.Commands = []cli.Command{
		{
			Name:   "daemon",
			Usage:  "start host daemon",
			Action: runDaemon,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "log-dir",
					Value: conf.ArgFetch("log-dir", ""),
					Usage: "directory to store job logs",
				},
				stateFlag,
				repoFlag,
			},
		},
		{
			Name:   "init",
			Usage:  "initialize host options",
			Action: cmd.Init,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "log-dir",
					Value: conf.ArgFetch("log-dir", "/var/log/pebblescape"),
					Usage: "directory to store job logs",
				},
				cli.StringFlag{
					Name:  "external-ip",
					Usage: "external IP address of host, defaults to the first IPv4 address of eth0",
				},
				stateFlag,
				repoFlag,
			},
		},
		{
			Name:   "receive",
			Usage:  "receive git push",
			Action: cmd.Receive,
		},
	}

	app.Run(os.Args)
}

func runDaemon(c *cli.Context) {
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
	}

	state := NewState(stateFile)

	if err := serveHTTP(state, gitRepos); err != nil {
		log.Fatal(err)
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	backend, err := NewDockerBackend(state, client)
	if err != nil {
		log.Fatal(err)
	}

	state.Restore(backend)

	<-make(chan error)
}
