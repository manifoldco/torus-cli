// Package cmd contains all of the Arigato cli commands
package cmd

import "github.com/urfave/cli"

// Cmds is the list of all cli commands
var Cmds []cli.Command

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return "Usage:\n" + spacer + ctx.App.HelpName + " " + ctx.Command.Name + " [command options] " + ctx.Command.ArgsUsage
}
