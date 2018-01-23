package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prompts"
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

	pe, cname, err := determinePath(ctx, args[0])
	if err != nil {
		return errs.NewErrorExitError("Could not unset credential", err)
	}

	name := args[0]
	if cname != nil {
		name = *cname
	}

	preamble := fmt.Sprintf("You are about to unset \"%s/%s\". This cannot be undone.", pe.String(), name)

	success, err := prompts.Confirm(nil, &preamble, true, true)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve confirmation", err)
	}
	if !success {
		return errs.ErrAbort
	}

	makers := valueMakers{}
	makers[name] = func() *apitypes.CredentialValue {
		return apitypes.NewUnsetCredentialValue()
	}

	s, p := spinner(fmt.Sprintf("Attempting to unset credential %s", name))
	s.Start()
	_, err = setCredentials(ctx, pe, makers, p)
	if err != nil {
		return errs.NewErrorExitError("Could not unset credential", err)
	}
	s.Stop()

	output := fmt.Sprintf("\nCredential %s has been unset at %s/%s.", name, pe, name)
	fmt.Println(output)

	return nil
}
