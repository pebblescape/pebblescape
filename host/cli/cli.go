package cli

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/flynn/go-docopt"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

type command struct {
	usage string
	f     interface{}
}

var commands = make(map[string]*command)

func Register(cmd string, f interface{}, usage string) *command {
	c := &command{usage: strings.TrimLeftFunc(usage, unicode.IsSpace), f: f}
	commands[cmd] = c
	return c
}

var ErrInvalidCommand = errors.New("invalid command")

func Run(name string, args []string) error {
	argv := make([]string, 1, 1+len(args))
	argv[0] = name
	argv = append(argv, args...)

	cmd, ok := commands[name]
	if !ok {
		return ErrInvalidCommand
	}
	parsedArgs, err := docopt.Parse(cmd.usage, argv, true, "", strings.Contains(cmd.usage, "[--]"))
	if err != nil {
		return err
	}

	switch f := cmd.f.(type) {
	case func(*docopt.Args, *docker.Client) error:
		client, err := docker.NewClientFromEnv()
		if err != nil {
			return err
		}
		return f(parsedArgs, client)
	case func(*docopt.Args):
		f(parsedArgs)
		return nil
	case func(*docopt.Args) error:
		return f(parsedArgs)
	case func() error:
		return f()
	case func():
		f()
		return nil
	}

	return fmt.Errorf("unexpected command type %T", cmd.f)
}
