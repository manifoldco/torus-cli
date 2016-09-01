package invites

import "github.com/urfave/cli"

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return "Usage:\n" + spacer + ctx.App.HelpName + " " + ctx.Command.Name + " [command options] " + ctx.Command.ArgsUsage
}
