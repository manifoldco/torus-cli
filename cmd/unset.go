package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/errs"
)

func init() {
	unset := cli.Command{
		Name:      "unset",
		Usage:     "Remove a secret from a service and environment",
		ArgsUsage: "<name|path>",
		Category:  "SECRETS",
		Flags:     append(setUnsetFlags, stdAutoAcceptFlag),
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

	pathexp, cname, err := determineCredential(ctx, args[0])
	if err != nil {
		return errs.NewErrorExitError("Could not unset credential", err)
	}

	preamble := fmt.Sprintf("You are about to unset \"%s/%s\". This cannot be undone.", pathexp.String(), *cname)

	abortErr := ConfirmDialogue(ctx, nil, &preamble)
	if abortErr != nil {
		return abortErr
	}

	var cred *apitypes.CredentialEnvelope
	cred, err = setCredential(ctx, args[0], func() *apitypes.CredentialValue {
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
