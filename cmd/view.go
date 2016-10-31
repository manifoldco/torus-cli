package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
)

func init() {
	view := cli.Command{
		Name:     "view",
		Usage:    "View secrets for the current service and environment",
		Category: "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),
			stdInstanceFlag,
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "list the sources of the values",
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
	if verbose {
		fmt.Printf("Credential path: %s\n\n", path)
	}
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	for _, secret := range secrets {
		value := (*secret.Body).GetValue()
		name := (*secret.Body).GetName()
		key := strings.ToUpper(name)
		if verbose {
			spath := (*secret.Body).GetPathExp().String() + "/" + name
			fmt.Fprintf(w, "%s=%s\t%s\n", key, value.String(), spath)
		} else {
			fmt.Fprintf(w, "%s=%s\n", key, value.String())

		}
	}
	w.Flush()
	return nil
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
		return nil, "", errs.NewErrorExitError("Error fetching user details", err)
	}

	parts := []string{
		"", ctx.String("org"), ctx.String("project"), ctx.String("environment"),
		ctx.String("service"), session.Username(), ctx.String("instance"),
	}

	path := strings.Join(parts, "/")

	secrets, err := client.Credentials.Get(c, path)
	if err != nil {
		return nil, "", errs.NewErrorExitError("Error fetching secrets", err)
	}

	cset := credentialSet{}
	for _, c := range secrets {
		cset.Add(c)
	}

	return cset.ToSlice(), path, nil
}
