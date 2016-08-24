// Package cmd contains all of the Arigato cli commands
package cmd

import "github.com/urfave/cli"

// Cmds is the list of all cli commands
var Cmds []cli.Command

// TODO indicates that a command or subcommand is not implemented.
func TODO(ctx *cli.Context) error {
	return cli.NewExitError("TODO: This command is not yet implemented", -1)
}
