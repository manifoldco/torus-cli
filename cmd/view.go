package cmd

import (
	"context"
	"encoding/json"
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

var formatFlag = newPlaceholder("format, f", "FORMAT", "Format used to display data (json, env, verbose)",
	"env", "TORUS_FORMAT", false)

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
			formatFlag,
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

	if ctx.Bool("verbose") && ctx.IsSet("format") {
		return errs.NewUsageExitError(
			"Cannot specify --format and --verbose at the same time", ctx)
	}

	format := ctx.String("format")
	if ctx.Bool("verbose") {
		format = "verbose"
	}

	switch format {
	case "env":
		return printEnvFormat(secrets, path)
	case "verbose":
		return printVerboseFormat(secrets, path)
	case "json":
		return printJSONFormat(secrets, path)
	default:
		return errs.NewUsageExitError("Unknown format: "+format, ctx)
	}
}

func printEnvFormat(secrets []apitypes.CredentialEnvelope, path string) error {
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	for _, secret := range secrets {
		value := (*secret.Body).GetValue()
		name := (*secret.Body).GetName()
		key := strings.ToUpper(name)
		fmt.Fprintf(w, "%s=%s\n", key, value.String())
	}
	w.Flush()

	return nil
}

func printVerboseFormat(secrets []apitypes.CredentialEnvelope, path string) error {
	fmt.Printf("Credential path: %s\n\n", path)

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	for _, secret := range secrets {
		value := (*secret.Body).GetValue()
		name := (*secret.Body).GetName()
		key := strings.ToUpper(name)
		spath := (*secret.Body).GetPathExp().String() + "/" + name
		fmt.Fprintf(w, "%s=%s\t%s\n", key, value.String(), spath)
	}
	w.Flush()

	return nil

}

func printJSONFormat(secrets []apitypes.CredentialEnvelope, path string) error {
	keyMap := make(map[string]interface{})

	for _, secret := range secrets {
		value := (*secret.Body).GetValue()
		name := (*secret.Body).GetName()
		v, err := value.Raw()
		if err != nil {
			return err
		}

		keyMap[name] = v
	}

	str, err := json.MarshalIndent(keyMap, "", "  ")
	if err != nil {
		return errs.NewErrorExitError("Could not marshal to json", err)
	}

	fmt.Printf("%s\n", str)
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
		return nil, "", err
	}

	identity, err := deriveIdentity(ctx, session)
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
