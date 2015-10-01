package cmd

import (
	"fmt"
	"log"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/client"
	"github.com/pebblescape/pebblescape/pkg/table"
)

func init() {
	RegisterCommand(cli.Command{
		Name:   "jobs",
		Usage:  "List jobs",
		Action: listJobs,
		Before: setClientPort,
	})

	RegisterCommand(cli.Command{
		Name:   "jobs:info",
		Usage:  "Show job info",
		Action: showJob,
		Before: setClientPort,
	})
}

func listJobs(c *cli.Context) {
	jobs, err := client.ListJobs()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("=== Jobs")
	for _, j := range jobs {
		fmt.Println(j.ID)
	}
}

func showJob(c *cli.Context) {
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "jobs:info")
		log.Fatal("You must specify a name")
	}

	job, err := client.GetJob(name)
	if err != nil {
		log.Fatalln(err)
	}

	tbl := table.New(2)
	tbl.Add("Status:", fmt.Sprintf("%v", job.Status))

	fmt.Println("=== " + job.ID)
	fmt.Print(tbl.String())
}
