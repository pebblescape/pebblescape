package cli

import (
	"fmt"

	"github.com/pebblescape/pebblescape/pkg/version"
)

func init() {
	Register("version", runVersion, `
usage: pebbles-host version
Show current version`)
}

func runVersion() {
	fmt.Println(version.String())
}
