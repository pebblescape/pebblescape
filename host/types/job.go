package host

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

type Job struct {
	Resurrect bool `json:"resurrect,omitempty"`
	*docker.Container
}
