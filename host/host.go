package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/flynn/go-docopt"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/cli"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/gitreceived"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
	"github.com/pebblescape/pebblescape/pkg/version"
)

const configFile = "/etc/pebblescape/host.json"

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

	cli.Register("daemon", runDaemon, `
usage: pebbles-host daemon [options]
options:
  --external-ip=IP      external IP of host
  --state=PATH          path to state file [default: /var/lib/pebblescape/host-state.bolt]
  --log-dir=DIR         directory to store job logs [default: /var/log/pebblescape]
  --git-port=PORT       port to listen for Git pushes on [default: 22]
  --repo-path=PATH      path for Git repo caches [default: /tmp/repos]
  --git-skip-auth=BOOL  disable Git client authentication [default: false]
  --git-keys=PATH       pem file containing private keys (read from SSH_PRIVATE_KEYS env by default)
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

	args, _ := docopt.Parse(usage, nil, true, version.String(), true)
	cmd := args.String["<command>"]
	cmdArgs := args.All["<args>"].([]string)

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

func runDaemon(args *docopt.Args, client *docker.Client) error {
	stateFile := args.String["--state"]
	gitKeyPath := args.String["--git-keys"]
	gitPort := args.String["--git-port"]
	gitRepos := args.String["--repo-path"]
	gitSkipAuth := args.Bool["--git-skip-auth"]

	var keys []byte
	if keyEnv := os.Getenv("SSH_PRIVATE_KEYS"); keyEnv != "" {
		keys = []byte(keyEnv)
	} else {
		pemBytes, err := ioutil.ReadFile(gitKeyPath)
		if err != nil {
			return errors.New("Failed to load private keys")
		}

		keys = pemBytes
	}

	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	receiver := path + " receive"

	state := NewState(stateFile)

	backend, err := NewDockerBackend(state, client)
	if err != nil {
		return err
	}

	cnt := backend.Inspect("f1e47b3a3dd2a6d8059d98233e3cd85f16dbd9604887f6b284a47df0582fafb2")
	log.Printf("%s", cnt.Config.Entrypoint)
	log.Printf("%s", cnt.Config.Cmd)

	log.Println("Starting git receiver")
	err = gitreceived.Serve(gitPort, gitRepos, gitSkipAuth, keys, receiver)
	if err != nil {
		return err
	}

	return <-make(chan error)
}
