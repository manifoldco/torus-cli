package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/identity"
)

const orgInviteFailed = "Could not send invitation to org, please try again."

func invitesSend(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) < 1 || args[0] == "" {
		return errs.NewUsageExitError("Missing email", ctx)
	}
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}
	email := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	org, err := client.Orgs.GetByName(context.Background(), ctx.String("org"))
	if err != nil {
		return errs.NewExitError(orgInviteFailed)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	// Identify the user attempting the command
	session, err := client.Session.Who(context.Background())
	if err != nil {
		return errs.NewExitError(orgInviteFailed)
	}

	// Retrieve teams for our target org
	teams, err := client.Teams.GetByOrg(context.Background(), org.ID)
	if err != nil {
		return errs.NewExitError(orgInviteFailed)
	}

	matchTeams := ctx.StringSlice("team")

	// ensure that even with custom teams, users are always invited to the
	// member team
	const memberTeam = "member"
	memberFound := false
	for _, team := range matchTeams {
		if team == memberTeam {
			memberFound = true
			break
		}
	}
	if !memberFound {
		matchTeams = append(matchTeams, memberTeam)
	}

	// Verify all team names supplied exist for this org
	teamIDs := make([]identity.ID, len(matchTeams))
	var missingTeams []string

TeamSearch:
	for i, teamName := range matchTeams {
		for _, team := range teams {
			if team.Body.Name == teamName {
				teamIDs[i] = *team.ID
				continue TeamSearch
			}
		}
		missingTeams = append(missingTeams, teamName)
	}

	// One of the supplied teams is not known to this org
	if len(missingTeams) > 0 {
		missingTeamNames := strings.Join(missingTeams, ", ")
		return errs.NewExitError("Unknown team(s): " + missingTeamNames)
	}
	if len(teamIDs) < 1 {
		return errs.NewExitError(orgInviteFailed)
	}

	err = client.OrgInvites.Send(context.Background(), email, *org.ID, *session.ID(), teamIDs)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError(email + " has already been invited to the " + org.Body.Name + " org")
		}
		return errs.NewExitError(orgInviteFailed)
	}

	fmt.Println("Invitation to join the " + org.Body.Name + " organization has been sent to " + email + ".")
	fmt.Println("\nThey will be added to the following teams once their invite has been confirmed:")
	fmt.Println("\n\t" + strings.Join(matchTeams, "\n\t"))
	fmt.Println("\nThey will receive an e-mail with instructions.")

	hints.Display([]string{"invites approve", "teams members"})
	return nil
}
