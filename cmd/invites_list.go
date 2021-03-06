package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/ui"
)

func invitesList(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	var states []string
	if ctx.Bool("approved") {
		states = []string{"approved"}
	} else {
		states = []string{"pending", "associated", "accepted"}
	}

	invites, err := client.OrgInvites.List(context.Background(), org.ID, states, "")
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve invites, please try again.", err)
	}

	if len(invites) < 1 {
		fmt.Println("No invites found.")
		return nil
	}

	inviteUserIDs := make(map[identity.ID]bool)
	for _, invite := range invites {
		if invite.Body.InviteeID != nil {
			inviteUserIDs[*invite.Body.InviteeID] = true
		}
		if invite.Body.ApproverID != nil {
			inviteUserIDs[*invite.Body.ApproverID] = true
		}
		inviteUserIDs[*invite.Body.InviterID] = true
	}

	var profileIDs []identity.ID
	for id := range inviteUserIDs {
		profileIDs = append(profileIDs, id)
	}

	// Lookup profiles of those who were invited
	profiles, err := client.Profiles.ListByID(context.Background(), profileIDs)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve invites, please try again.", err)
	}

	nameByID := make(map[string]string)
	usernameByID := make(map[string]string)
	for _, profile := range profiles {
		nameByID[profile.ID.String()] = profile.Body.Name
		usernameByID[profile.ID.String()] = profile.Body.Username
	}

	fmt.Println("")
	if ctx.Bool("approved") {
		fmt.Println("Listing approved invitations for org " + org.Body.Name)
	} else {
		fmt.Println("Listing all pending and accepted invitations for org " + org.Body.Name)
	}
	fmt.Println("")

	w := ansiterm.NewTabWriter(os.Stdout, 2, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", ui.BoldString("Invited E-Mail"), ui.BoldString("Name"), ui.BoldString("Username"), ui.BoldString("State"), ui.BoldString("Invited by"), ui.BoldString("Creation Date"))
	for _, invite := range invites {
		inviter := nameByID[invite.Body.InviterID.String()]
		if inviter == "" {
			continue
		}
		identity := invite.Body.Email
		inviteeName := "-"
		inviteeUsername := "-"
		if invite.Body.InviteeID != nil {
			inviteeName = nameByID[invite.Body.InviteeID.String()]
			inviteeUsername = usernameByID[invite.Body.InviteeID.String()]
		}
		var state string
		switch invite.Body.State {
		case "pending":
			state = ui.FaintString("awaiting acceptance")
		case "accepted":
			state = ui.ColorString(ui.Default, "awaiting approval")
		case "approved":
			state = ui.ColorString(ui.Green, "approved")
		default:
			state = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", identity, inviteeName, ui.FaintString(inviteeUsername), state, inviter, invite.Body.Created.Format(time.RFC3339))

	}
	w.Flush()
	fmt.Println("")

	hints.Display(hints.InvitesApprove, hints.OrgMembers)
	return nil
}
