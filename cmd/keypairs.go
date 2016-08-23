package cmd

import "github.com/urfave/cli"

func init() {
	keypairs := cli.Command{
		Name:     "keypairs",
		Usage:    "View and generate organization keypairs",
		Category: "ORGANIZATIONS",
	}
	Cmds = append(Cmds, keypairs)
}
