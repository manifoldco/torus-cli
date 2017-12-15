package cmd

import (
	"context"
	"fmt"
	"os"
	//"text/tabwriter"
	"time"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
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

	var org *envelope.Org
	var orgs []envelope.Org

	// Retrieve the org name supplied via the --org flag.
	// This flag is optional. If none was supplied, then
	// orgFlagArgument will be set to "". In this case,
	// prompt the user to select an org.
	orgName := ctx.String("org")

	if orgName == "" {
		// Retrieve list of available orgs
		orgs, err = client.Orgs.List(c)
		if err != nil {
			return errs.NewExitError("Failed to retrieve orgs list.")
		}

		// Prompt user to select from list of existing orgs
		idx, _, err := SelectExistingOrgPrompt(orgs)
		if err != nil {
			return errs.NewErrorExitError("Failed to select org.", err)
		}

		org = &orgs[idx]

	} else {
		// If org flag was used, identify the org supplied.
		org, err = client.Orgs.GetByName(c, orgName)
		if err != nil {
			return errs.NewErrorExitError("Failed to retrieve org " + orgName, err)
		}
		if org == nil {
			return errs.NewExitError("org " + orgName + " not found.")
		}
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

	usernameByID := make(map[string]string)
	for _, profile := range profiles {
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
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ui.Bold("E-Mail"), ui.Bold("Username"), ui.Bold("State"), ui.Bold("Invited by"), ui.Bold("Creation Date"))
	for _, invite := range invites {
		inviter := usernameByID[invite.Body.InviterID.String()]
		if inviter == "" {
			continue
		}
		identity := invite.Body.Email
		invitee := "-"
		if invite.Body.InviteeID != nil {
			invitee = usernameByID[invite.Body.InviteeID.String()]
		}
		var state string
		switch invite.Body.State {
		case "pending":
			state = ui.Color(ansiterm.DarkGray, "awaiting acceptance")
		case "accepted":
			state = ui.Color(ansiterm.Yellow, "awaiting approval")
		case "approved":
			state = ui.Color(ansiterm.Green, "approved")
		default:
			state = "-"
		}

		fmt.Fprintln(w, identity+"\t"+invitee+"\t"+state+"\t"+inviter+"\t"+invite.Body.Created.Format(time.RFC3339))
	}
	w.Flush()
	fmt.Println("")

	hints.Display(hints.InvitesApprove, hints.OrgMembers)
	return nil
}
