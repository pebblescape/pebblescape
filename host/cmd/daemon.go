package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/pebblescape/pebblescape/host/api"
	"github.com/pebblescape/pebblescape/host/config"
	"github.com/pebblescape/pebblescape/host/http"
	"github.com/pebblescape/pebblescape/pkg/shutdown"
)

func init() {
	RegisterCommand(cli.Command{
		Name:   "daemon",
		Usage:  "Start host daemon",
		Action: startDaemon,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "port, p",
				Value: "4592",
				Usage: "Pebblescape port",
			},
			cli.StringFlag{
				Name:  "log-dir",
				Value: "",
				Usage: "directory to store daemon logs",
			},
			cli.StringFlag{
				Name:  "host-key",
				Value: "",
				Usage: "host authentication key",
			},
			cli.StringFlag{
				Name:  "home",
				Value: "/var/lib/pebblescape",
				Usage: "Pebblescape home dir",
			},
			cli.BoolFlag{
				Name:  "dev",
				Usage: "development mode",
			},
		},
	})
}

func startDaemon(c *cli.Context) {
	defer shutdown.Exit()

	port := c.String("port")
	dev := c.Bool("dev")
	logger := setupLogger(c.String("log-dir"))

	conf, err := config.Ensure(config.ConfigFile, c.String("host-key"), c.String("home"))
	if err != nil {
		if os.IsExist(err) {
			log.Fatal("Host is already running")
		}

		log.Fatal(err)
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	hostAPI := api.New(client, conf, logger)

	shutdown.BeforeExit(func() {
		log.Println("Cleaning up...")
		if err := hostAPI.StopDB(); err != nil {
			log.Printf("Error stopping host DB: %v", err)
		}

		if err := conf.Cleanup(); err != nil {
			log.Printf("Error cleaning config: %v", err)
		}
	})

	if dev {
		log.Printf("Starting host, dev mode\n")
	} else {
		log.Printf("Starting host\n")
	}

	if err := hostAPI.StartDB(dev); err != nil {
		log.Printf("Error starting host DB: %v", err)
		return
	}

	if err := http.Serve(port, hostAPI, conf, logger); err != nil {
		log.Printf("HTTP API error: %v", err)
	}
}

func setupLogger(logDir string) *log.Logger {
	var logger *log.Logger

	log.SetFlags(log.Ldate | log.Lmicroseconds)

	if logDir != "" {
		logFile := filepath.Join(logDir, "host.log")
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			log.Fatalf("could not not mkdir for logs: %s", err)
		}

		hostlog, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer hostlog.Close()

		log.Printf("Logging to %s\n", logFile)
		log.SetOutput(io.MultiWriter(hostlog, os.Stdout))
		logger = log.New(io.MultiWriter(hostlog, os.Stderr), "", log.Flags())
	} else {
		logger = log.New(os.Stderr, "", log.Flags())
	}

	return logger
}
