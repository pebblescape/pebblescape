package cmd

import (
	"log"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/client"
)

func init() {
	cmd := cli.Command{
		Name:   "users",
		Usage:  "list users",
		Action: ListUsers,
	}

	RegisterCommand(cmd)
}

func ListUsers(c *cli.Context) {
	users, err := client.ListUsers()
	if err != nil {
		log.Fatalln(err)
	}

	for _, u := range users {
		log.Println(u.Name)
	}
}
