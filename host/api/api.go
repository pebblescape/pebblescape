package api

import (
	"errors"
	"log"
	"path/filepath"
	"time"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	_ "github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/lib/pq" // import postgres driver
	"github.com/pebblescape/pebblescape/host/config"
)

// API is an interface that binds together all parts of the host infrastructure.
type API struct {
	Client *docker.Client
	Config *config.Config
	Logger *log.Logger
	DB     *sqlx.DB
}

// New creates and returns a new host API.
func New(client *docker.Client, conf *config.Config, logger *log.Logger) *API {
	return &API{client, conf, logger, nil}
}

// StartDB boots up a host database instance.
func (a *API) StartDB(dev bool) error {
	dbPath := filepath.Join(a.Config.Home, "db")

	_, err := a.Client.InspectImage("postgres")
	if err != nil {
		if err == docker.ErrNoSuchImage {
			a.Logger.Println("Pulling DB image")
			if err := a.Client.PullImage(docker.PullImageOptions{
				Repository: "postgres",
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
			"label": {"com.pebblescape.db=true"},
		},
	})
	if err != nil {
		return err
	}

	for _, c := range results {
		a.Config.DbID = c.ID
		if err := a.StopDB(); err != nil {
			return err
		}
	}

	a.Config.DbPass = "pebblescapeia"

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: "postgres",
			Volumes: map[string]struct{}{
				"/var/lib/postgresql/data/pgdata": {},
			},
			Labels: map[string]string{
				"com.pebblescape.db": "true",
			},
			Env: []string{
				"POSTGRES_PASSWORD=" + a.Config.DbPass,
				"PGDATA=/var/lib/postgresql/data/pgdata",
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				dbPath + ":/var/lib/postgresql/data/pgdata:rw",
			},
		},
	}

	if dev {
		opts.Config.ExposedPorts = map[docker.Port]struct{}{
			"5432/tcp": {},
		}
		opts.HostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{
				HostPort: "4593",
			}},
		}
		opts.HostConfig.Binds = []string{
			"/tmp/pebblescapedb:/var/lib/postgresql/data/pgdata:rw",
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

	// Wait for DB to initialize
	time.Sleep(3 * time.Second)
	if err := a.ConnectDB(); err != nil {
		return err
	}

	return migrateDB(a.DB)
}

// StopDB cleanly shuts down running host database instance.
func (a *API) StopDB() error {
	if a.DB != nil {
		a.DB.Close()
	}

	a.Client.StopContainer(a.Config.DbID, 5)
	a.Client.WaitContainer(a.Config.DbID)
	a.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    a.Config.DbID,
		Force: true,
	})

	return nil
}

// ConnectDB creates a new sqlx connection to running host database instance.
func (a *API) ConnectDB() error {
	cnt, err := a.Client.InspectContainer(a.Config.DbID)
	if err != nil {
		return err
	}

	port, ok := cnt.NetworkSettings.Ports["5432/tcp"]
	if !ok {
		return errors.New("DB port not exposed")
	}

	db, err := sqlx.Connect("postgres", "port="+port[0].HostPort+" password="+a.Config.DbPass+" user=postgres dbname=postgres sslmode=disable")
	if err != nil {
		return err
	}

	a.DB = db

	return nil
}
