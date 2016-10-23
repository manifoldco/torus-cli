package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/errs"
	"github.com/arigatomachine/cli/identity"
)

func invitesList(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	org, err := client.Orgs.GetByName(context.Background(), ctx.String("org"))
	if err != nil {
		return errs.NewExitError("Could not retrieve org information.")
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	var states []string
	if ctx.Bool("approved") {
		states = []string{"approved"}
	} else {
		states = []string{"pending", "associated", "accepted"}
	}

	invites, err := client.Invites.List(context.Background(), org.ID, states)
	if err != nil {
		return errs.NewExitError("Failed to retrieve invites, please try again.")
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
		return errs.NewExitError("Failed to retrieve invites, please try again.")
	}

	usernameByID := make(map[string]string)
	for _, profile := range *profiles {
		usernameByID[profile.ID.String()] = profile.Body.Username
	}

	fmt.Println("")
	if ctx.Bool("approved") {
		fmt.Println("Listing approved invitations for the " + ctx.String("org") + " org")
	} else {
		fmt.Println("Listing all pending and accepted invitations for the " + ctx.String("org") + " org")
	}
	fmt.Println("")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, "EMAIL\tUSERNAME\tSTATE\tINVITED BY\tCREATION DATE")
	fmt.Fprintln(w, " \t \t \t ")
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
		fmt.Fprintln(w, identity+"\t"+invitee+"\t"+invite.Body.State+"\t"+inviter+"\t"+invite.Body.Created.Format(time.RFC3339))
	}
	w.Flush()
	fmt.Println("")

	return nil
}
