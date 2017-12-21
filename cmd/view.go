package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"github.com/juju/ansiterm"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/ui"
)

func init() {
	view := cli.Command{
		Name:     "view",
		Usage:    "View secrets for the current service and environment",
		Category: "SECRETS",
		Flags: []cli.Flag{
			orgFlag("Use this organization.", false),
			projectFlag("Use this project.", false),
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),
			userFlag("Use this user.", false),
			machineFlag("Use this machine.", false),
			stdInstanceFlag,
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Lists the sources of the secrets (shortcut for --format verbose)",
			},
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, checkRequiredFlags, viewCmd,
		),
	}

	Cmds = append(Cmds, view)
}

func viewCmd(ctx *cli.Context) error {
	secrets, path, err := getSecrets(ctx)
	if err != nil {
		return err
	}

	verbose := ctx.Bool("verbose")

	w := os.Stdout

	fmt.Fprintf(w, "Credential path: %s\n\n", path)

	tw := ansiterm.NewTabWriter(w, 2, 0, 2, ' ', 0)
	for _, secret := range secrets {
		value := (*secret.Body).GetValue().String()
		name := (*secret.Body).GetName()
		spath := (*secret.Body).GetPathExp().String() + "/" + name

		if verbose {
			if strings.Contains(value, " ") {
				fmt.Fprintf(tw, "%s\t=\t%q\t(%s)\n", ui.Bold(name), value, ui.Faint(spath))
			} else {
				fmt.Fprintf(tw, "%s\t=\t%s\t(%s)\n", ui.Bold(name), value, ui.Faint(spath))
			}
		} else {
			if strings.Contains(value, " ") {
				fmt.Fprintf(tw, "%s\t=\t%q\n", ui.Bold(name), value)
			} else {
				fmt.Fprintf(tw, "%s\t=\t%s\n", ui.Bold(name), value)
			}
		}
	}

	return tw.Flush()

	hints.Display(hints.Link, hints.Run, hints.Export)

	return err
}

func getSecrets(ctx *cli.Context) ([]apitypes.CredentialEnvelope, string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, "", err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	session, err := client.Session.Who(c)
	if err != nil {
		return nil, "", err
	}

	identity, err := deriveIdentity(ctx, session)
	if err != nil {
		return nil, "", err
	}

	org, err := getOrgWithPrompt(client, c, ctx.String("org"))
	if err != nil {
		return nil, "", err
	}

	project, err := getProjectWithPrompt(client, c, org, ctx.String("project"))
	if err != nil {
		return nil, "", err
	}

	parts := []string{
		"", org.Body.Name, project.Body.Name, ctx.String("environment"),
		ctx.String("service"), identity, ctx.String("instance"),
	}

	path := strings.Join(parts, "/")

	secrets, err := client.Credentials.Get(c, path)
	if err != nil {
		return nil, "", errs.NewErrorExitError("Error fetching secrets", err)
	}

	cset := credentialSet{}
	for _, c := range secrets {
		if err := cset.Add(c); err != nil {
			return nil, "", errs.NewErrorExitError("Error compacting secrets", err)
		}
	}

	return cset.ToSlice(), path, nil
}
