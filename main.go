package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/cmd"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"
)

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		cmd.VersionLookup(ctx)
	}

	preferences, _ := prefs.NewPreferences(true)
	ui.Init(preferences)

	app := cli.NewApp()
	app.Version = config.Version
	app.Usage = "A secure, shared workspace for secrets"
	app.Commands = cmd.Cmds
	app.Run(os.Args)
}
