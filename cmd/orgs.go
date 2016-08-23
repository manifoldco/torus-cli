package cmd

import "github.com/urfave/cli"

func init() {
	orgs := cli.Command{
		Name:     "orgs",
		Usage:    "View and create organizations",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, orgs)
}
