package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap"
)

func init() {
	bootstrap := cli.Command{
		Name:     "bootstrap",
		Usage:    "Bootstrap a new machine using Torus Gatekeeper",
		Category: "SYSTEM",
		Flags: []cli.Flag{
			authProviderFlag("Auth provider for bootstrapping", true),
			urlFlag("Gatekeeper URL for bootstrapping", true),
			orgFlag("Org the machine will belong to", false),
			roleFlag("Role the machine will belong to", false),
		},
		Action: chain(checkRequiredFlags, bootstrapCmd),
	}

	Cmds = append(Cmds, bootstrap)
}

// bootstrapCmd is the cli.Command for Bootstrapping machine configuration from the Gatekeeper
func bootstrapCmd(ctx *cli.Context) error {
	cloud := bootstrap.Type(ctx.String("auth"))

	provider, err := bootstrap.New(cloud)
	if err != nil {
		fmt.Printf("Bootstrap failed: %s\n", err)
	}

	return provider.Bootstrap(
		ctx.String("url"),
		ctx.String("org"),
		ctx.String("role"),
	)
}
