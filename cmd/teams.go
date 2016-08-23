package cmd

import "github.com/urfave/cli"

func init() {
	teams := cli.Command{
		Name:     "teams",
		Usage:    "View and manipulate teams within an organization",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, teams)
}
