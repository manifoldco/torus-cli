package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"
)

func init() {
	prefs := cli.Command{
		Name:     "prefs",
		Usage:    "Manage tool preferences",
		Category: "SYSTEM",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "Set the preference key to the value. If value is omitted, then it sets it to true",
				ArgsUsage: "<key> <value>",
				Action: func(ctx *cli.Context) error {
					return setPref(ctx)
				},
			},
			{
				Name:  "list",
				Usage: "Show your account preferences",
				Action: func(ctx *cli.Context) error {
					return listPref(ctx)
				},
			},
		},
	}
	Cmds = append(Cmds, prefs)
}

func listSection(name string, count int, section interface{}) error {
	if count <= 0 {
		return nil
	}

	fmt.Println(name)
	fc := ini.Empty()
	err := ini.ReflectFrom(fc, section)
	if err != nil {
		return err
	}

	_, err = fc.WriteTo(ui.Child(4))
	return err
}

func listPref(ctx *cli.Context) error {
	const loadErr = "Failed to load prefs."
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return errs.NewErrorExitError(loadErr, err)
	}

	filepath, _ := prefs.RcPath()
	fmt.Println("\n" + filepath + "\n")

	coreCount := preferences.CountFields("Core")
	err = listSection("[core]", coreCount, &preferences.Core)
	if err != nil {
		return errs.NewErrorExitError(loadErr, err)
	}

	defaultsCount := preferences.CountFields("Defaults")
	err = listSection("[defaults]", defaultsCount, &preferences.Defaults)
	if err != nil {
		return errs.NewErrorExitError(loadErr, err)
	}

	if defaultsCount < 1 && coreCount < 1 {
		fmt.Println("No preferences set. Use 'torus prefs set' to update.")
		fmt.Println("")
	}

	return nil
}

func setPref(ctx *cli.Context) error {
	args := ctx.Args()
	key := args.Get(0)
	value := args.Get(1)
	if len(key) < 1 || len(value) < 1 {
		return errs.NewUsageExitError("Must supply a key and value", ctx)
	}

	if len(strings.Split(key, ".")) < 2 {
		return errs.NewExitError("Key must be have at least two dot delimited segments.")
	}

	return setPrefByName(key, value)
}

func setPrefByName(key, value string) error {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return errs.NewErrorExitError("Failed to load prefs.", err)
	}

	// Validate public key file
	if key == "core.public_key_file" {
		err := prefs.ValidatePublicKey(value)
		if err != nil {
			return err
		}

		value, err = filepath.Abs(value)
		if err != nil {
			return err
		}
	}

	// Set value inside prefs struct
	result, err := preferences.SetValue(key, value)
	if err != nil {
		return err
	}

	// Reflect struct to ini format
	cfg := ini.Empty()
	err = ini.ReflectFrom(cfg, &result)
	if err != nil {
		return errs.NewErrorExitError("Failed to save preferences.", err)
	}

	// Save updated ini to filePath
	rcPath, _ := prefs.RcPath()
	err = cfg.SaveTo(rcPath)
	if err != nil {
		return errs.NewErrorExitError("Failed to save preferences.", err)
	}

	fmt.Println("Preferences updated.")
	return nil
}
