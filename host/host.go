package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/docopt/docopt-go"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/cli"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
	"github.com/pebblescape/pebblescape/pkg/version"
)

const configFile = "/etc/pebblescape/host.json"

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	cli.Register("daemon", runDaemon, `
usage: pebbles-host daemon [options]
options:
  --external-ip=IP       external IP of host
  --state=PATH           path to state file [default: /var/lib/pebblescape/host-state.bolt]
  --log-dir=DIR          directory to store job logs [default: /var/log/pebblescape]
	`)
}

func main() {
	defer shutdown.Exit()

	usage := `usage: pebbles-host [-h|--help] [--version] <command> [<args>...]
Options:
  -h, --help                 Show this message
  --version                  Show current version
Commands:
  help                       Show usage for a specific command
  init                       Create cluster configuration for daemon
  daemon                     Start the daemon
  update                     Update Pebblescape components
  version                    Show current version
See 'pebbles-host help <command>' for more information on a specific command.
`

	args, _ := docopt.Parse(usage, nil, true, version.Version, true)
	cmd := args["<command>"].(string)
	cmdArgs := args["<args>"].([]string)

	if cmd == "help" {
		if len(cmdArgs) == 0 { // `pebbles-host help`
			fmt.Println(usage)
			return
		} else { // `pebbles-host help <command>`
			cmd = cmdArgs[0]
			cmdArgs = []string{"--help"}
		}
	}

	if cmd == "daemon" {
		// merge in args and env from config file, if available
		var c *config.Config
		var err error

		c, err = config.Open(configFile)
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("error opening config file %s: %s", configFile, err)
		}
		if c == nil {
			c = &config.Config{}
		}

		cmdArgs = append(cmdArgs, c.Args...)
		for k, v := range c.Env {
			os.Setenv(k, v)
		}
	}

	if err := cli.Run(cmd, cmdArgs); err != nil {
		if err == cli.ErrInvalidCommand {
			fmt.Printf("ERROR: %q is not a valid command\n\n", cmd)
			fmt.Println(usage)
			shutdown.ExitWithCode(1)
		}
		shutdown.Fatal(err)
	}
}

func runDaemon(args map[string]interface{}, client *docker.Client) error {
	fmt.Printf("%v", args)
	<-make(chan struct{})
	return nil
}
