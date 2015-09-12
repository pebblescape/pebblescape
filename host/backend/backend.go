package backend

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/types"
)

type Backend interface {
	Inspect(string) *docker.Container
	Run(*host.Job) error
	// Start(string) error
	// Cleanup([]string) error
	// UnmarshalState(map[string]*host.Job, map[string][]byte) error
}
