package cmd

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/primitive"
)

func init() {
	deny := cli.Command{
		Name:      "deny",
		Usage:     "Decrease access given to a team or role by creating and attaching a new policy",
		ArgsUsage: "<crudl> <path> <team|machine-role>",
		Category:  "ACCESS CONTROL",
		Action:    chain(ensureDaemon, ensureSession, denyCmd),
	}

	Cmds = append(Cmds, deny)
}

func denyCmd(ctx *cli.Context) error {
	return doCrudl(ctx, primitive.PolicyEffectDeny, 0x0)
}
