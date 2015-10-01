package host

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

type Backend interface {
	Inspect(string) *docker.Container
	Run(*Job) error
	Restore(*Job) error
	Cleanup() error
}
