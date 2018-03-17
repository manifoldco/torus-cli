package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/pathexp"
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

	fmt.Fprintf(w, "Credential path: %s\n\n", displayPathExp(path))

	tw := ansiterm.NewTabWriter(w, 2, 0, 2, ' ', 0)
	for _, secret := range secrets {
		value := (*secret.Body).GetValue().String()
		name := (*secret.Body).GetName()
		spath := displayPathExp((*secret.Body).GetPathExp()) + "/" + name

		if verbose {
			if strings.Contains(value, " ") {
				fmt.Fprintf(tw, "%s\t=\t%q\t(%s)\n", ui.BoldString(name), value, ui.FaintString(spath))
			} else {
				fmt.Fprintf(tw, "%s\t=\t%s\t(%s)\n", ui.BoldString(name), value, ui.FaintString(spath))
			}
		} else {
			if strings.Contains(value, " ") {
				fmt.Fprintf(tw, "%s\t=\t%q\n", ui.BoldString(name), value)
			} else {
				fmt.Fprintf(tw, "%s\t=\t%s\n", ui.BoldString(name), value)
			}
		}
	}

	tw.Flush()

	hints.Display(hints.Link, hints.Run, hints.Export)

	return nil
}

func getSecrets(ctx *cli.Context) ([]apitypes.CredentialEnvelope, *pathexp.PathExp, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	session, err := client.Session.Who(c)
	if err != nil {
		return nil, nil, err
	}

	identity := deriveIdentity(session)
	path, err := deriveExplicitPathExp(ctx.String("org"), ctx.String("project"),
		ctx.String("environment"), ctx.String("service"), identity)
	if err != nil {
		return nil, nil, errs.NewErrorExitError("Error deriving credential path", err)
	}

	s, p := spinner("Decrypting credentials")
	s.Start()
	defer s.Stop()
	secrets, err := client.Credentials.Get(c, path.String(), p)
	if err != nil {
		return nil, nil, errs.NewErrorExitError("Error fetching secrets", err)
	}

	cset := credentialSet{}
	for _, c := range secrets {
		if err := cset.Add(c); err != nil {
			return nil, nil, errs.NewErrorExitError("Error compacting secrets", err)
		}
	}

	out, err := pathexp.New(ctx.String("org"), ctx.String("project"),
		[]string{ctx.String("environment")}, []string{ctx.String("service")},
		[]string{"*"}, []string{"*"})
	if err != nil {
		return nil, nil, err
	}

	return cset.ToSlice(), out, nil
}

func deriveExplicitPathExp(org, project, env, service, identity string) (*pathexp.PathExp, error) {
	return pathexp.New(org, project, []string{env}, []string{service}, []string{identity}, []string{"1"})
}
