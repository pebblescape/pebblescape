package backend

import (
	"fmt"
	"sync"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/types"
)

const (
	dockerNamespace = "com.pebblescape."
)

type DockerBackend struct {
	client *docker.Client
	state  host.State

	containersMtx sync.RWMutex
	containers    map[string]*dockerContainer
}

type dockerContainer struct {
	job  *host.Job
	c    *docker.Container
	b    *DockerBackend
	done chan struct{}
}

func NewDockerBackend(client *docker.Client, state host.State) (host.Backend, error) {
	return &DockerBackend{
		client:     client,
		state:      state,
		containers: make(map[string]*dockerContainer),
	}, nil
}

func (b *DockerBackend) Run(j *host.Job) error {
	cnt, err := b.client.CreateContainer(docker.CreateContainerOptions{
		Config:     j.Config,
		HostConfig: j.HostConfig,
	})
	if err != nil {
		return err
	}

	b.state.SetContainerID(j.ID, cnt.ID)

	container := &dockerContainer{
		job:  j,
		b:    b,
		c:    cnt,
		done: make(chan struct{}),
	}

	go container.watch()

	err = b.client.StartContainer(j.ContainerID, nil)
	if err != nil {
		return err
	}

	return nil
}

func (b *DockerBackend) Restore(j *host.Job) error {
	cnt, err := b.client.InspectContainer(j.ContainerID)
	if err != nil {
		return b.Run(j)
	}

	if cnt.State.Running {
		container := &dockerContainer{
			job:  j,
			b:    b,
			c:    cnt,
			done: make(chan struct{}),
		}

		go container.watch()

		return nil
	}

	b.client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    j.ContainerID,
		Force: true,
	})

	return b.Run(j)
}

func (b *DockerBackend) Inspect(id string) *docker.Container {
	cnt, _ := b.client.InspectContainer(id)
	return cnt
}

func (b *DockerBackend) Cleanup() error {
	return nil
}

func (c *dockerContainer) watch() error {
	events := make(chan *docker.APIEvents)

	defer func() {
		c.b.containersMtx.Lock()
		delete(c.b.containers, c.job.ID)
		c.b.containersMtx.Unlock()
		close(c.done)

		c.b.client.RemoveEventListener(events)
	}()

	c.b.containersMtx.Lock()
	c.b.containers[c.job.ContainerID] = c
	c.b.containersMtx.Unlock()

	if err := c.b.client.AddEventListener(events); err != nil {
		return err
	}

	for event := range events {
		if event.ID != c.job.ContainerID {
			continue
		}
		fmt.Printf("%v\n", event)

		switch event.Status {
		case "destroy":
			close(events)
		}
	}
	c.b.state.RemoveJob(c.job.ID)
	// c.l.state.SetStatusFailed(c.job.ID, errors.New("unknown failure"))

	return nil
}
