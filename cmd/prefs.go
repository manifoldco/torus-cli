package cmd

import "github.com/urfave/cli"

func init() {
	prefs := cli.Command{
		Name:     "prefs",
		Usage:    "View and set preferences",
		Category: "ACCOUNT",
	}
	Cmds = append(Cmds, prefs)
}
