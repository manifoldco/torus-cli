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
			userFlag("Use this user.", false),
			machineFlag("Use this machine.", false),
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

func deriveCurrentIdentity(c context.Context, ctx *cli.Context, client *api.Client) (string, error) {
	if ctx.String("user") != "" && ctx.String("machine") != "" {
		return "", errs.NewExitError(
			"You can only supply --user or --machine, not both.")
	}

	identity := ""
	if ctx.String("user") != "" {
		identity = ctx.String("user")
	}

	var err error
	if ctx.String("machine") != "" {
		identity, err = identityString("machine", ctx.String("machine"))
		if err != nil {
			return "", err
		}
	}

	if identity == "" {
		session, err := client.Session.Who(c)
		if err != nil {
			return "", err
		}

		identity = session.Username()
		if session.Type() == "machine" {
			identity = "machine-" + identity
		}
	}

	return identity, nil
}

func getSecrets(ctx *cli.Context) ([]apitypes.CredentialEnvelope, string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, "", err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	identity, err := deriveCurrentIdentity(c, ctx, client)
	if err != nil {
		return nil, "", err
	}

	parts := []string{
		"", ctx.String("org"), ctx.String("project"), ctx.String("environment"),
		ctx.String("service"), identity, ctx.String("instance"),
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
