package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/config"
)

func init() {
	view := cli.Command{
		Name:     "view",
		Usage:    "View secrets for the current service and environment",
		Category: "SECRETS",
		Flags: []cli.Flag{
			StdOrgFlag,
			StdProjectFlag,
			StdEnvFlag,
			ServiceFlag("Use this service.", "default", true),
			StdInstanceFlag,
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "list the sources of the values",
			},
		},
		Action: Chain(
			EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
			SetUserEnv, checkRequiredFlags, viewCmd,
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
		value := secret.Body.Value
		if value.IsUnset() {
			continue
		}
		key := strings.ToUpper(secret.Body.Name)
		if verbose {
			spath := secret.Body.PathExp + "/" + secret.Body.Name
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

	self, err := client.Users.Self(c)
	if err != nil {
		return nil, "", cli.NewExitError("Error fetching user details: "+err.Error(), -1)
	}

	parts := []string{
		"", ctx.String("org"), ctx.String("project"), ctx.String("environment"),
		ctx.String("service"), self.Body.Username, ctx.String("instance"),
	}

	path := strings.Join(parts, "/")

	secrets, err := client.Credentials.Get(c, path)
	if err != nil {
		return nil, "", cli.NewExitError("Error fetching secrets: "+err.Error(), -1)
	}

	return secrets, path, nil
}
