package cmd

import (
	"fmt"
	"log"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/pkg/table"
)

func init() {
	RegisterCommand(cli.Command{
		Name:   "apps",
		Usage:  "List apps",
		Action: apps,
	})

	RegisterCommand(cli.Command{
		Name:   "apps:create",
		Usage:  "Create app",
		Action: createApp,
	})
}

func apps(c *cli.Context) {
	api := getApi()

	apps, err := api.Apps()
	if err != nil {
		log.Fatal(err)
	}

	tbl := table.New(1)

	for _, app := range apps {
		tbl.Add(app.Name)
	}

	fmt.Println("=== Apps")
	fmt.Print(tbl.String())
}

func createApp(c *cli.Context) {
	name := c.Args()[0]
	if name == "" {
		log.Fatal("Must specify name")
	}

	// api := getApi()
}
