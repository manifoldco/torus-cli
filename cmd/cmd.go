// Package cmd contains all of the Torus cli commands
package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
)

// Cmds is the list of all cli commands
var Cmds []cli.Command

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return "Usage:\n" + spacer + ctx.App.HelpName + " " + ctx.Command.Name + " [command options] " + ctx.Command.ArgsUsage
}

var progress api.ProgressFunc = func(evt *api.Event, err error) {
	if evt != nil {
		fmt.Println(evt.Message)
	}
}
