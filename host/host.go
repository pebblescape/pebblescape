package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/cmd"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/gitreceived"
	"github.com/pebblescape/pebblescape/pkg/paths"
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
	gitPortFlag := cli.IntFlag{
		Name:  "git-port",
		Value: conf.ArgFetchI("git-port", 22),
		Usage: "port to listen for Git pushes on",
	}
	repoFlag := cli.StringFlag{
		Name:  "repo-path",
		Value: conf.ArgFetch("repo-path", "/tmp/repos"),
		Usage: "path for Git repo caches",
	}
	skipAuthFlag := cli.BoolFlag{
		Name:   "git-skip-auth",
		Usage:  "disable Git client authentication",
		EnvVar: "PEBBLES_GIT_SKIP_AUTH",
	}
	keysFlag := cli.StringFlag{
		Name:   "git-keys",
		Usage:  "pem file containing private keys",
		EnvVar: "PEBBLES_GIT_KEYS",
	}

	app.Commands = []cli.Command{
		{
			Name:   "daemon",
			Usage:  "start host daemon",
			Action: runDaemon,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "log-dir",
					Usage: "directory to store job logs",
				},
				stateFlag,
				gitPortFlag,
				repoFlag,
				skipAuthFlag,
				keysFlag,
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
				gitPortFlag,
				repoFlag,
				skipAuthFlag,
				keysFlag,
			},
		},
	}

	app.Run(os.Args)
}

func runDaemon(c *cli.Context) {
	logDir := c.String("log-dir")
	stateFile := c.String("state")
	gitKeys := c.String("git-keys")
	gitPort := c.String("git-port")
	gitRepos := c.String("repo-path")
	gitSkipAuth := c.Bool("git-skip-auth")

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
		log.SetOutput(io.MultiWriter(hostlog, os.Stdout))
		log.Printf("Logging to %s\n", logFile)
	}

	var keys []byte
	if _, err := os.Stat(gitKeys); err == nil {
		pemBytes, err := ioutil.ReadFile(gitKeys)
		if err != nil {
			log.Fatal("Failed to load private keys")
		}

		keys = pemBytes
	} else {
		keys = []byte(os.Getenv("SSH_PRIVATE_KEYS"))
	}

	state := NewState(stateFile)

	if err := serveHTTP(state); err != nil {
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

	receiver := paths.SelfPath() + " receive"
	err = gitreceived.Serve(gitPort, gitRepos, gitSkipAuth, keys, receiver)
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan error)
}
