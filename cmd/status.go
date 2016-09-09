package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/prefs"
	"github.com/urfave/cli"
)

func init() {
	status := cli.Command{
		Name:     "status",
		Usage:    "Show the current Arigato status associated with your account and project",
		Category: "CONTEXT",
		Flags: []cli.Flag{
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),

			// These flags are hidden so we can still parse out the values
			// from the prefs files and env vars, but we don't display
			// them to the users in help.
			// A user could still set the flag on the command line though :(
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "org", EnvVar: "AG_ORG", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "project", EnvVar: "AG_PROJECT", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "instance", EnvVar: "AG_INSTANCE", Value: "*", Hidden: true},
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
		return cli.NewExitError(msg, -1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	self, err := client.Users.Self(c)
	if err != nil {
		return cli.NewExitError("Error fetching user details: "+err.Error(), -1)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Identity:\t%s <%s>\n", self.Body.Name, self.Body.Email)
	fmt.Fprintf(w, "Username:\t%s\n\n", self.Body.Username)
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

	parts := []string{"", org, project, env, service, self.Body.Username, instance}
	credPath := strings.Join(parts, "/")
	fmt.Printf("\nCredential path: %s\n", credPath)

	return nil
}
