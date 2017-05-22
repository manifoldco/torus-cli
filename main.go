package main

import (
	"log"
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

	// For command line usage, we hide any log messages; our regular error
	// flow will catch them. Logging is only used with the daemon.
	log.SetOutput(devnull{})

	preferences, _ := prefs.NewPreferences()
	ui.Init(preferences)

	app := cli.NewApp()
	app.Name = "torus"
	app.HelpName = "torus"
	app.Usage = "A secure, shared workspace for secrets"
	app.Version = config.Version
	app.Commands = cmd.Cmds
	app.Run(os.Args)
}

// devnull swallows up all log output.
type devnull struct{}

func (devnull) Write(p []byte) (n int, err error) { return len(p), nil }
