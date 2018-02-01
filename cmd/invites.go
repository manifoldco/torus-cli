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
					orgFlag("org to invite user to", true),
					newSlicePlaceholder("team, t", "TEAM", "team to add user to", "member", "", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setSliceDefaults, checkRequiredFlags, invitesSend,
				),
			},
			{
				Name:  "list",
				Usage: "List outstanding invitations for an organization. These invites have yet to be approved.",
				Flags: []cli.Flag{
					orgFlag("org to list invites for", false),
					cli.BoolFlag{
						Name:  "approved, a",
						Usage: "Show only approved invites",
					},
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, invitesList,
				),
			},
			{
				Name:      "approve",
				Usage:     "Approve an invitation previously sent to an email address to join an organization",
				ArgsUsage: "<email>",
				Flags: []cli.Flag{
					orgFlag("org to approve invite for", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs,
					loadPrefDefaults, checkRequiredFlags, invitesApprove,
				),
			},
			{
				Name:      "accept",
				Usage:     "Accept an invitation to join an organization",
				ArgsUsage: "<email> <code>",
				Flags: []cli.Flag{
					orgFlag("org to accept invite for", true),
				},
				Action: chain(
					ensureDaemon, loadDirPrefs,
					loadPrefDefaults, checkRequiredFlags, invitesAccept,
				),
			},
		},
	}
	Cmds = append(Cmds, invites)
}
