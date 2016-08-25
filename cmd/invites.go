package cmd

import "github.com/urfave/cli"

func init() {
	invites := cli.Command{
		Name:     "invites",
		Usage:    "View and accept organization invites",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, invites)
}
