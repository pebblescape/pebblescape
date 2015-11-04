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
	repo := getApi().AppsRepo()

	apps, err := repo.List()
	if err != nil {
		log.Fatal(err)
	}

	tbl := table.New(2)

	for _, app := range apps {
		tbl.Add(app.Name, app.ID)
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
