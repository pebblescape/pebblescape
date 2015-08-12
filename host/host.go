package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	// "github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/cmd"
	// "github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/gitreceived"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
	"github.com/pebblescape/pebblescape/pkg/version"
)

const configFile = "/etc/pebblescape/host.json"

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)
}

func main() {
	defer shutdown.Exit()

	app := cli.NewApp()
	app.Name = "pebblescape"
	app.Usage = "manage pebblescape host"
	app.Version = version.String()

	app.Commands = []cli.Command{
		{
			Name:   "daemon",
			Usage:  "start host daemon",
			Action: runDaemon,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "state",
					Value: "/var/lib/pebblescape/host-state.bolt",
					Usage: "path to state file",
				},
				cli.StringFlag{
					Name:  "log-dir",
					Value: "/var/log/pebblescape",
					Usage: "directory to store job logs",
				},
				cli.IntFlag{
					Name:  "git-port",
					Value: 22,
					Usage: "port to listen for Git pushes on",
				},
				cli.StringFlag{
					Name:  "repo-path",
					Value: "/tmp/repos",
					Usage: "path for Git repo caches",
				},
				cli.BoolFlag{
					Name:  "git-skip-auth",
					Usage: "disable Git client authentication",
				},
				cli.StringFlag{
					Name:   "git-keys",
					Usage:  "pem file containing private keys",
					EnvVar: "SSH_PRIVATE_KEYS",
				},
			},
		},
		{
			Name:   "init",
			Usage:  "initialize host options",
			Action: cmd.Init,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "external-ip",
					Usage: "external IP address of host, defaults to the first IPv4 address of eth0",
				},
				cli.StringFlag{
					Name:  "file",
					Value: "/etc/pebblescape/host.json",
					Usage: "file to write to",
				},
				cli.StringFlag{
					Name:  "state",
					Value: "/var/lib/pebblescape/host-state.bolt",
					Usage: "path to state file",
				},
				cli.StringFlag{
					Name:  "log-dir",
					Value: "/var/log/pebblescape",
					Usage: "directory to store job logs",
				},
				cli.IntFlag{
					Name:  "git-port",
					Value: 22,
					Usage: "port to listen for Git pushes on",
				},
				cli.StringFlag{
					Name:  "repo-path",
					Value: "/tmp/repos",
					Usage: "path for Git repo caches",
				},
				cli.BoolFlag{
					Name:  "git-skip-auth",
					Usage: "disable Git client authentication",
				},
				cli.StringFlag{
					Name:   "git-keys",
					Usage:  "pem file containing private keys",
					EnvVar: "SSH_PRIVATE_KEYS",
				},
			},
		},
	}

	app.Run(os.Args)
}

func runDaemon(c *cli.Context) {
	stateFile := c.String("state")
	gitKeys := c.String("git-keys")
	gitPort := c.String("git-port")
	gitRepos := c.String("repo-path")
	gitSkipAuth := c.Bool("git-skip-auth")

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

	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	receiver := path + " receive"

	// _, err := docker.NewClientFromEnv()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	state := NewState(stateFile)

	// backend, err := NewDockerBackend(state, client)
	// if err != nil {
	// 	return err
	// }

	if err := serveHTTP(state); err != nil {
		log.Fatal(err)
	}

	err = gitreceived.Serve(gitPort, gitRepos, gitSkipAuth, keys, receiver)
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan error)
}
