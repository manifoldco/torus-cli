package cmd

import "github.com/urfave/cli"

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
					OrgFlag("org to invite user to", true),
					newSlicePlaceholder("team, t", "TEAM", "team to add user to", nil),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, invitesSend,
				),
			},
			{
				Name:      "list",
				Usage:     "List outstanding invitations for an organization. These invites have yet to be approved.",
				ArgsUsage: "",
				Flags: []cli.Flag{
					OrgFlag("org to list invites for", true),
					cli.BoolFlag{
						Name:  "approved",
						Usage: "Show only approved invites",
					},
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, invitesList,
				),
			},
			{
				Name:      "approve",
				Usage:     "Approve an invitation previously sent to an email address to join an organization",
				ArgsUsage: "",
				Flags: []cli.Flag{
					OrgFlag("org to approve invite for", true),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs,
					LoadPrefDefaults, SetUserEnv, checkRequiredFlags, invitesApprove,
				),
			},
			{
				Name:      "accept",
				Usage:     "Accept an invitation to join an organization",
				ArgsUsage: "<email> <code>",
				Flags: []cli.Flag{
					OrgFlag("org to approve invite for", true),
				},
				Action: Chain(
					EnsureDaemon, LoadDirPrefs,
					LoadPrefDefaults, SetUserEnv, checkRequiredFlags, invitesAccept,
				),
			},
		},
	}
	Cmds = append(Cmds, invites)
}
