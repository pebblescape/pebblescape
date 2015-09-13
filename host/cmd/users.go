package cmd

import (
	"fmt"
	"log"

	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/pebblescape/pebblescape/host/client"
	"github.com/pebblescape/pebblescape/host/types"
	"github.com/pebblescape/pebblescape/pkg/table"
)

func init() {
	RegisterCommand(cli.Command{
		Name:   "users",
		Usage:  "List users",
		Action: listUsers,
		Before: setClientPort,
	})

	RegisterCommand(cli.Command{
		Name:   "users:info",
		Usage:  "Show user info",
		Action: showUser,
		Before: setClientPort,
	})

	RegisterCommand(cli.Command{
		Name:   "users:create",
		Usage:  "Create user",
		Action: createUser,
		Before: setClientPort,
	})
}

func listUsers(c *cli.Context) {
	users, err := client.ListUsers()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("=== Users")
	for _, u := range users {
		fmt.Println(u.Name)
	}
}

func showUser(c *cli.Context) {
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "users:info")
		log.Fatal("You must specify a name")
	}

	user, err := client.GetUser(name)
	if err != nil {
		log.Fatalln(err)
	}

	tbl := table.New(2)
	tbl.Add("Token:", user.Token)

	fmt.Println("=== " + user.Name)
	fmt.Print(tbl.String())
}

func createUser(c *cli.Context) {
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, "users:create")
		log.Fatal("You must specify a name")
	}

	user := host.User{Name: name}

	err := client.CreateUser(user)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Created user " + user.Name)
}
