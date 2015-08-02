package cli

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/flynn/go-docopt"
	"github.com/pebblescape/pebblescape/host/config"
)

func init() {
	Register("init", runInit, `
usage: pebbles-host init [options]
options:
	--external-ip=IP    external IP address of host, defaults to the first IPv4 address of eth0
	--file=NAME         file to write to [default: /etc/pebblescape/host.json]
	`)
}

func runInit(args *docopt.Args) error {
	c := config.New()

	if ip := args.String["--external-ip"]; ip != "" {
		c.Args = append(c.Args, "--external-ip", ip)
	}

	return c.WriteTo(args.String["--file"])
}
