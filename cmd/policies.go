package cmd

import "github.com/urfave/cli"

func init() {
	policies := cli.Command{
		Name:     "policies",
		Usage:    "View and manipulate access control list policies",
		Category: "ACCESS CONTROL",
	}
	Cmds = append(Cmds, policies)
}
