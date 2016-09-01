package cmd

import (
	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/cmd/invites"
)

func init() {
	invites := cli.Command{
		Name:     "invites",
		Usage:    "View and accept organization invites",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:      "send",
				Usage:     "Send an invitation to join an organization to an email address",
				ArgsUsage: "<email>",
				Flags: []cli.Flag{
					OrgFlag("org to invite user to"),
					newSlicePlaceholder("team, t", "TEAM", "team to add user to", nil),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs,
					LoadPrefDefaults, SetUserEnv, invites.Send,
				),
			},
			{
				Name:      "list",
				Usage:     "List outstanding invitations for an organization. These invites have yet to be approved.",
				ArgsUsage: "",
				Flags: []cli.Flag{
					OrgFlag("org to list invites for"),
					cli.BoolFlag{
						Name:  "approved",
						Usage: "Show only approved invites",
					},
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs,
					LoadPrefDefaults, SetUserEnv, invites.List,
				),
			},
		},
	}
	Cmds = append(Cmds, invites)
}
