package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/errs"
)

func init() {
	unset := cli.Command{
		Name:      "unset",
		Usage:     "Remove a secret from a service and environment",
		ArgsUsage: "<name|path>",
		Category:  "SECRETS",
		Flags:     setUnsetFlags,
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setSliceDefaults, unsetCmd,
		),
	}

	Cmds = append(Cmds, unset)
}

func unsetCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 1 {
		msg := "Name or path is required."
		if len(args) > 1 {
			msg = "Too many arguments provided."
		}
		return errs.NewUsageExitError(msg, ctx)
	}

	cred, err := setCredential(ctx, args[0], func() *apitypes.CredentialValue {
		return apitypes.NewUnsetCredentialValue()
	})

	if err != nil {
		return errs.NewErrorExitError("Could not unset credential", err)
	}

	name := (*cred.Body).GetName()
	pe := (*cred.Body).GetPathExp()
	output := fmt.Sprintf("\nCredential %s has been unset at %s/%s.", name, pe, name)
	fmt.Println(output)

	return nil
}
