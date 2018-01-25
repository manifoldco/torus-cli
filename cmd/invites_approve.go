package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
)

const approveInviteFailed = "Could not approve invitation to org, please try again."

func invitesApprove(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) < 1 {
		return errs.NewUsageExitError("Missing email", ctx)
	}
	email := ctx.Args()[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	org, err := client.Orgs.GetByName(context.Background(), ctx.String("org"))
	if err != nil {
		return errs.NewExitError(approveInviteFailed)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	states := []string{"accepted"}
	invites, err := client.OrgInvites.List(context.Background(), org.ID, states, "")
	if err != nil {
		return errs.NewExitError("Failed to retrieve invites, please try again.")
	}

	// Find the target invite id
	var targetInvite *identity.ID
	for _, invite := range invites {
		if invite.Body.Email == email {
			targetInvite = invite.ID
		}
	}
	if targetInvite == nil {
		return errs.NewExitError("Invite not found.")
	}

	s, p := spinner(fmt.Sprintf("Attempting to approve invite for %s", email))
	s.Start()
	err = client.OrgInvites.Approve(context.Background(), *targetInvite, p)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("You have approved " + email + "'s invitation.")
	fmt.Println("")
	fmt.Println("They are now a member of the organization!")

	return nil
}
