package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arigatomachine/cli/prefs"

	"github.com/go-ini/ini"
	"github.com/kr/text"
	"github.com/urfave/cli"
)

func init() {
	prefs := cli.Command{
		Name:     "prefs",
		Usage:    "View and set preferences",
		Category: "ACCOUNT",
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

func listPref(ctx *cli.Context) error {
	preferences, err := prefs.NewPreferences(false)
	if err != nil {
		return cli.NewExitError("Failed to load prefs: "+err.Error(), -1)
	}

	spacer := "    "
	filepath, _ := prefs.RcPath()
	fmt.Println("\n" + filepath + "\n")

	coreCount := preferences.CountFields("Core")
	defaultsCount := preferences.CountFields("Defaults")

	if coreCount > 0 {
		fmt.Println("[core]")
		fc := ini.Empty()
		err = ini.ReflectFrom(fc, &preferences.Core)
		fc.WriteToIndent(text.NewIndentWriter(os.Stdout, []byte(spacer)), spacer)
	}

	if defaultsCount > 0 {
		fmt.Println("[defaults]")
		fd := ini.Empty()
		err = ini.ReflectFrom(fd, &preferences.Defaults)
		fd.WriteToIndent(text.NewIndentWriter(os.Stdout, []byte(spacer)), spacer)
	}

	if defaultsCount < 1 && coreCount < 1 {
		fmt.Println("No preferences set. Use 'torus prefs set' to update.")
		fmt.Println("")
	}

	return nil
}

func setPref(ctx *cli.Context) error {
	preferencess, err := prefs.NewPreferences(false)
	if err != nil {
		return cli.NewExitError("Failed to load prefs: "+err.Error(), -1)
	}

	args := ctx.Args()
	key := args.Get(0)
	value := args.Get(1)
	if len(key) < 1 || len(value) < 1 {
		text := "error: must supply a key and value\n\n" + usageString(ctx)
		return cli.NewExitError(text, -1)
	}

	if len(strings.Split(key, ".")) < 2 {
		text := "error: key must be have at least two dot delimited segments"
		return cli.NewExitError(text, -1)
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
	result, err := preferencess.SetValue(key, value)
	if err != nil {
		return err
	}

	// Reflect struct to ini format
	cfg := ini.Empty()
	err = ini.ReflectFrom(cfg, &result)
	if err != nil {
		fmt.Println(err.Error())
		return cli.NewExitError("error: failed to save preferences", -1)
	}

	// Save updated ini to filePath
	rcPath, _ := prefs.RcPath()
	err = cfg.SaveTo(rcPath)
	if err != nil {
		fmt.Println(err.Error())
		return cli.NewExitError("error: failed to save preferences", -1)
	}

	fmt.Println("Preferences updated.")
	return nil
}
