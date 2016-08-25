package cmd

import "github.com/urfave/cli"

func init() {
	services := cli.Command{
		Name:     "services",
		Usage:    "View and manipulate services within an organization",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, services)
}
