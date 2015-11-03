package api

import (
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fzzy/radix/redis"
	"github.com/pebblescape/pebblescape/host/config"
)

type Api struct {
	Client *docker.Client
	Config *config.Config
	Logger *log.Logger
	DB     *redis.Client
}

func New(client *docker.Client, conf *config.Config, logger *log.Logger) *Api {
	return &Api{client, conf, logger, nil}
}

func (a *Api) StartDb(dev bool) error {
	dbPath := filepath.Join(a.Config.Home, "db")

	_, err := a.Client.InspectImage("redis")
	if err != nil {
		if err == docker.ErrNoSuchImage {
			a.Logger.Println("Pulling redis image")
			if err := a.Client.PullImage(docker.PullImageOptions{
				Repository: "redis",
				Tag:        "latest",
			}, docker.AuthConfiguration{}); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	results, err := a.Client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": []string{"com.pebblescape.db=true"},
		},
	})
	if err != nil {
		return err
	}

	for _, c := range results {
		a.Config.DbID = c.ID
		if err := a.StopDb(); err != nil {
			return err
		}
	}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: "redis",
			Volumes: map[string]struct{}{
				"/data": {},
			},
			Cmd: []string{
				"redis-server",
				"--appendonly",
				"yes",
			},
			Labels: map[string]string{
				"com.pebblescape.db": "true",
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				dbPath + ":/data:rw",
			},
		},
	}

	if dev {
		opts.Config.ExposedPorts = map[docker.Port]struct{}{
			"6379/tcp": {},
		}
		opts.HostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
			"6379/tcp": []docker.PortBinding{docker.PortBinding{
				HostPort: "4593",
			}},
		}
	} else {
		opts.HostConfig.PublishAllPorts = true
	}

	cnt, err := a.Client.CreateContainer(opts)
	if err != nil {
		return err
	}

	err = a.Client.StartContainer(cnt.ID, nil)
	if err != nil {
		return err
	}

	a.Config.DbID = cnt.ID
	if err := a.Config.Write(); err != nil {
		return err
	}

	return a.ConnectDb()
}

func (a *Api) StopDb() error {
	a.Client.StopContainer(a.Config.DbID, 5)
	a.Client.WaitContainer(a.Config.DbID)
	a.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    a.Config.DbID,
		Force: true,
	})

	return nil
}

func (a *Api) ConnectDb() error {
	cnt, err := a.Client.InspectContainer(a.Config.DbID)
	if err != nil {
		return err
	}

	port, ok := cnt.NetworkSettings.Ports["6379/tcp"]
	if !ok {
		return errors.New("DB port not exposed")
	}

	rds, err := redis.Dial("tcp", "localhost:"+port[0].HostPort)
	if err != nil {
		return err
	}

	test := "lucha!lucha!"
	echo, err := rds.Cmd("ECHO", test).Str()
	if err != nil {
		return err
	}

	if echo != test {
		return errors.New("DB echo does not match")
	}

	a.DB = rds

	return nil
}

func (a *Api) Apps() ([]*App, error) {
	path := filepath.Join(a.Config.Home, "apps")
	results := make([]*App, 0)

	test := &App{"yo"}
	PersistApp(filepath.Join(path, "yo.json"), test)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		app, _ := ParseApp(filepath.Join(path, f.Name()))

		results = append(results, app)
	}

	return results, nil
}
