package host

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

type Job struct {
	ID          string             `json:"id,omitempty"`
	ContainerID string             `json:"container_id,omitempty"`
	Status      JobStatus          `json:"status,omitempty"`
	Config      *docker.Config     `json:"config,omitempty"`
	HostConfig  *docker.HostConfig `json:"host_config,omitempty"`
}

type JobStatus uint8

func (s JobStatus) String() string {
	return map[JobStatus]string{
		StatusStarting: "starting",
		StatusRunning:  "running",
		StatusDone:     "done",
		StatusCrashed:  "crashed",
		StatusFailed:   "failed",
	}[s]
}

const (
	StatusStarting JobStatus = iota
	StatusRunning
	StatusDone
	StatusCrashed
	StatusFailed
)
