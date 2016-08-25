package cmd

import "github.com/urfave/cli"

func init() {
	envs := cli.Command{
		Name:     "envs",
		Usage:    "View and manipulate environments within an organization",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, envs)
}
