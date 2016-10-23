package cmd

import (
	"flag"
	"testing"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/errs"
	"github.com/arigatomachine/cli/prefs"
)

func TestChain(t *testing.T) {
	t.Run("aborts on first error", func(t *testing.T) {
		firstRan := false
		secondRan := false

		expected := errs.NewExitError("error")
		err := chain(
			func(ctx *cli.Context) error {
				firstRan = true
				if secondRan {
					t.Error("Second chained func ran first")
				}
				return expected
			},
			func(ctx *cli.Context) error {
				secondRan = true
				return nil
			},
		)(&cli.Context{})

		if err != expected {
			t.Error("Chain did not return first func's error")
		}

		if !firstRan {
			t.Error("First func did not run")
		}
		if secondRan {
			t.Error("Second func was run")
		}
	})

	t.Run("runs all chained funcs", func(t *testing.T) {
		firstRan := false
		secondRan := false

		chain(
			func(ctx *cli.Context) error {
				firstRan = true
				if secondRan {
					t.Error("Second chained func ran first")
				}
				return nil
			},
			func(ctx *cli.Context) error {
				secondRan = true
				return nil
			},
		)(&cli.Context{})

		if !(firstRan && secondRan) {
			t.Error("Both chained funcs did not run")
		}
	})
}

func TestReflectArgs(t *testing.T) {
	cmd := cli.Command{
		Flags: []cli.Flag{cli.StringFlag{Name: "org"}},
	}
	p := &prefs.Preferences{
		Core: prefs.Core{
			Context: true,
		},
		Defaults: prefs.Defaults{
			Organization: "org thing",
		},
	}

	t.Run("Exits early if core.context is false", func(t *testing.T) {
		p := &prefs.Preferences{
			Core: prefs.Core{
				Context: false,
			},
			Defaults: prefs.Defaults{
				Organization: "org thing",
			},
		}

		flagset := flag.NewFlagSet("", flag.ContinueOnError)
		flagset.String("org", "", "")
		ctx := cli.NewContext(nil, flagset, nil)
		ctx.Command = cmd
		err := reflectArgs(ctx, p, p.Defaults, "ini")
		if err != nil {
			t.Error("loadPrefDefaults errored: " + err.Error())
		}

		if ctx.IsSet("org") {
			t.Error("org argument should not have been set with context disabled")
		}
	})

	t.Run("Does not overwrite a set value", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ContinueOnError)
		flagset.String("org", "", "")
		ctx := cli.NewContext(nil, flagset, nil)
		ctx.Command = cmd
		ctx.Set("org", "good value")

		err := reflectArgs(ctx, p, p.Defaults, "ini")
		if err != nil {
			t.Error("loadPrefDefaults errored: " + err.Error())
		}

		if ctx.String("org") != "good value" {
			t.Error("loadPrefDefaults overwrote a set argument.")
		}
	})

	t.Run("Sets unset values", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ContinueOnError)
		flagset.String("org", "", "")
		ctx := cli.NewContext(nil, flagset, nil)
		ctx.Command = cmd

		err := reflectArgs(ctx, p, p.Defaults, "ini")
		if err != nil {
			t.Error("loadPrefDefaults errored: " + err.Error())
		}

		if ctx.String("org") != "org thing" {
			t.Error("loadPrefDefaults did not set argument")
		}
	})
}

func TestCheckRequiredFlags(t *testing.T) {
	app := cli.App{
		Name: "test",
	}
	cmd := cli.Command{
		Flags: []cli.Flag{
			orgFlag("an org", true),
			projectFlag("a project", false),
		},
	}

	t.Run("Unset non-required flags are ignored", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ContinueOnError)
		flagset.String("org", "my org", "")
		ctx := cli.NewContext(&app, flagset, nil)
		ctx.Command = cmd

		err := checkRequiredFlags(ctx)
		if err != nil {
			t.Error("unset non-required flag caused an error")
		}
	})

	t.Run("Unset required flags cause an error", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ContinueOnError)
		ctx := cli.NewContext(&app, flagset, nil)
		ctx.Command = cmd

		err := checkRequiredFlags(ctx)
		if err == nil {
			t.Error("unset required flag did not error")
		}
	})

	t.Run("Set required flags do not cause an error", func(t *testing.T) {
		flagset := flag.NewFlagSet("", flag.ContinueOnError)
		flagset.String("org", "my org", "")
		ctx := cli.NewContext(&app, flagset, nil)
		ctx.Command = cmd

		err := checkRequiredFlags(ctx)
		if err != nil {
			t.Error("set required flag caused an error")
		}
	})
}
