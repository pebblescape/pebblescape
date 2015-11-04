package cmd

import (
	"fmt"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/api"
	"github.com/pebblescape/pebblescape/pkg/table"
)

func init() {
	RegisterCommand(cli.Command{
		Name:   "apps",
		Usage:  "List apps",
		Action: apps,
		Before: setApi,
	})

	RegisterCommand(cli.Command{
		Name:   "apps:create",
		Usage:  "Create app",
		Action: createApp,
		Before: setApi,
	})
}

func apps(c *cli.Context) {
	repo := host.GetAppRepo()

	apps, err := repo.List()
	if err != nil {
		fatal(err)
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
		fatal("Must specify name")
	}

	app := &api.App{
		Name: name,
	}

	repo := host.GetAppRepo()
	if err := repo.Create(app); err != nil {
		fatal(err)
	}

	tbl := table.New(2)
	tbl.Add("ID:", app.ID)

	fmt.Println("=== App " + app.Name)
	fmt.Print(tbl.String())
}
