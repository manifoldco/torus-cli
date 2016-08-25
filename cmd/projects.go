package cmd

import "github.com/urfave/cli"

func init() {
	projects := cli.Command{
		Name:     "projects",
		Usage:    "View and manipulate projects within an organization",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, projects)
}
