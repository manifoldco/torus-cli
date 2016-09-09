package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/apitypes"
)

func init() {
	unset := cli.Command{
		Name:      "unset",
		Usage:     "Remove a secret from a service and environment",
		ArgsUsage: "<name|path>",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			newSlicePlaceholder("environment, e", "ENV", "Use this environment.",
				nil, "AG_ENVIRONMENT", true),
			newSlicePlaceholder("service, s", "SERVICE", "Use this service.",
				[]string{"default"}, "AG_SERVICE", true),
			newSlicePlaceholder("user, u", "USER", "Use this user (identity).",
				[]string{"*"}, "AG_USER", true),
			newSlicePlaceholder("instance, i", "INSTANCE", "Use this instance.",
				[]string{"1"}, "AG_INSTANCE", true),
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults, unsetCmd,
		),
	}

	Cmds = append(Cmds, unset)
}

func unsetCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 1 {
		msg := "name or path is required.\n"
		if len(args) > 1 {
			msg = "Too many arguments provided.\n"
		}
		msg += usageString(ctx)
		return cli.NewExitError(msg, -1)
	}

	cred, err := setCredential(ctx, args[0], func() *apitypes.CredentialValue {
		return apitypes.NewUnsetCredentialValue()
	})

	if err != nil {
		return cli.NewExitError("Could not unset credential: "+err.Error(), -1)
	}

	name := cred.Body.Name
	pe := cred.Body.PathExp
	fmt.Printf("\nCredential %s has been unset at %s/%s\n", name, pe, name)

	return nil
}
