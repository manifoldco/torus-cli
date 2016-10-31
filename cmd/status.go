package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"

	"github.com/urfave/cli"
)

func init() {
	status := cli.Command{
		Name:     "status",
		Usage:    "Show the current Torus status associated with your account and project",
		Category: "CONTEXT",
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
	preferences, err := prefs.NewPreferences(true)
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
		return errs.NewErrorExitError("Error fetching user details", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	if session.Type() == apitypes.MachineSession {
		fmt.Fprintf(w, "Machine ID:\t%s\n", session.ID())
		fmt.Fprintf(w, "Machine Token ID:\t%s\n", session.AuthID())
		fmt.Fprintf(w, "Machine Name:\t%s\n\n", session.Username())
	} else {
		fmt.Fprintf(w, "Identity:\t%s <%s>\n", session.Name(), session.Email())
		fmt.Fprintf(w, "Username:\t%s\n\n", session.Username())
	}

	w.Flush()

	err = checkRequiredFlags(ctx)
	if err != nil {
		fmt.Printf("You are not inside a linked working directory. "+
			"Use '%s link' to link your project.\n", ctx.App.Name)
		return nil
	}

	org := ctx.String("org")
	project := ctx.String("project")
	env := ctx.String("environment")
	service := ctx.String("service")
	instance := ctx.String("instance")

	fmt.Fprintf(w, "Org:\t%s\n", org)
	fmt.Fprintf(w, "Project:\t%s\n", project)
	fmt.Fprintf(w, "Environment:\t%s\n", env)
	fmt.Fprintf(w, "Service:\t%s\n", service)
	fmt.Fprintf(w, "Instance:\t%s\n", instance)
	w.Flush()

	parts := []string{"", org, project, env, service, session.Username(), instance}
	credPath := strings.Join(parts, "/")
	fmt.Printf("\nCredential path: %s\n", credPath)

	return nil
}
