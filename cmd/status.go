package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"

	"github.com/urfave/cli"
)

func init() {
	status := cli.Command{
		Name:     "status",
		Usage:    "Show the current Torus status associated with your account and project",
		Category: "PROJECT STRUCTURE",
		Flags: []cli.Flag{
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),

			// These flags are hidden so we can still parse out the values
			// from the prefs files and env vars, but we don't display
			// them to the users in help.
			// A user could still set the flag on the command line though :(
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "org", EnvVar: "TORUS_ORG", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "project", EnvVar: "TORUS_PROJECT", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "instance", EnvVar: "TORUS_INSTANCE", Value: "1", Hidden: true},
				Required:   true,
			},
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, statusCmd,
		),
	}

	Cmds = append(Cmds, status)
}

func statusCmd(ctx *cli.Context) error {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return err
	}

	if !preferences.Core.Context {
		msg := fmt.Sprintf("Context is disabled. Use '%s prefs' to enable it.", ctx.App.Name)
		return errs.NewExitError(msg)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	session, err := client.Session.Who(c)
	if err != nil {
		return errs.NewErrorExitError("Error fetching identity", err)
	}

	err = checkRequiredFlags(ctx)
	if err != nil {
		fmt.Printf("You are not inside a linked working directory. "+
			"Use '%s link' to link your project.\n", ctx.App.Name)
		return nil
	}

	identity, err := deriveIdentity(ctx, session)
	if err != nil {
		return err
	}

	org := ctx.String("org")
	project := ctx.String("project")
	env := ctx.String("environment")
	service := ctx.String("service")
	instance := ctx.String("instance")

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Org"), org)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Project"), project)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Environment"), env)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Service"), service)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Identity"), ui.FaintString(identity))
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Instance"), instance)
	w.Flush()

	parts := []string{"", org, project, env, service, identity, instance}
	credPath := strings.Join(parts, "/")
	fmt.Printf("\n%s: %s\n", ui.BoldString("Credential Path"), credPath)

	return nil
}
