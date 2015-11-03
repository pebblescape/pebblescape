package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func init() {
	RegisterCommand(cli.Command{
		Name:   "receive",
		Usage:  "Receive git push",
		Action: receive,
	})
}

func receive(c *cli.Context) {
	log.SetFlags(0)

	app := c.Args()[0]
	rev := c.Args()[1]
	cache := c.Args()[2]

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		cnt, err := buildApp(app, client, cache)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Built app %v rev %v into container %v\n", app, rev, cnt.ID)
	} else {
		log.Fatal("no app input received")
	}
}

func buildApp(app string, client *docker.Client, cacheRoot string) (*docker.Container, error) {
	cachePath := path.Join(cacheRoot, app)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return nil, fmt.Errorf("Could not not mkdir for cache: %s", err)
	}

	cnt, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:     "gliderlabs/herokuish",
			OpenStdin: true,
			Volumes: map[string]struct{}{
				"/tmp/cache": {},
			},
			Cmd: []string{
				"/bin/bash",
				"-c",
				"mkdir -p /app && tar -xC /app && /build",
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				cachePath + ":/tmp/cache:rw",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	err = client.StartContainer(cnt.ID, nil)
	if err != nil {
		return nil, err
	}

	err = client.AttachToContainer(docker.AttachToContainerOptions{
		Container:   cnt.ID,
		InputStream: os.Stdin,
		Stdin:       true,
		Stream:      true,
	})
	if err != nil {
		return nil, err
	}

	err = client.Logs(docker.LogsOptions{
		Container:    cnt.ID,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stdout,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
	})
	if err != nil {
		return nil, err
	}

	code, err := client.WaitContainer(cnt.ID)
	if err != nil {
		return nil, err
	}
	if code != 0 {
		return nil, errors.New("Container returned non-zero exit code")
	}

	return cnt, nil
}
