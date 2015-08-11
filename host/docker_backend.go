package main

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/types"
)

const (
	dockerNamespace = "com.pebblescape."
)

func NewDockerBackend(state *State, client *docker.Client) (Backend, error) {
	return &DockerBackend{
		state:      state,
		client:     client,
		containers: make(map[string]*docker.Container),
	}, nil
}

type DockerBackend struct {
	state      *State
	client     *docker.Client
	containers map[string]*docker.Container
}

func (b *DockerBackend) Run(j *host.Job) error {
	return nil
}

func (b *DockerBackend) Inspect(id string) *docker.Container {
	cnt, _ := b.client.InspectContainer(id)
	return cnt
}
