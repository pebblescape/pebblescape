package cmd

import (
	"errors"
	"log"
	"os"
	"path"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func init() {
	cmd := cli.Command{
		Name:   "receive",
		Usage:  "receive git push",
		Action: Receive,
	}

	RegisterCommand(cmd)
}

func Receive(c *cli.Context) {
	log.SetFlags(0)

	app := c.Args()[0]
	// rev := c.Args()[1]

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		cnt, err := buildApp(app, client)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Built app %s into container %s\n", app, cnt.ID)
	} else {
		log.Fatal("no app input received")
	}
}

func buildApp(app string, client *docker.Client) (*docker.Container, error) {
	cachePath := path.Join("/tmp/pebbles-cache", app)

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
