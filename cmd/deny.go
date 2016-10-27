package cmd

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/primitive"
)

func init() {
	deny := cli.Command{
		Name:      "deny",
		Usage:     "Deny a team permission to access specific resources",
		ArgsUsage: "<crudl> <path> <team>",
		Category:  "ACCESS CONTROL",
		Action:    chain(ensureDaemon, ensureSession, denyCmd),
	}

	Cmds = append(Cmds, deny)
}

func denyCmd(ctx *cli.Context) error {
	return doCrudl(ctx, primitive.PolicyEffectDeny, 0x0)
}
