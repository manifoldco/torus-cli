package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/urfave/cli"
	"github.com/juju/ansiterm"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/ui"
)

func init() {
	teams := cli.Command{
		Name:     "teams",
		Usage:    "Manage teams and their members",
		Category: "ACCESS CONTROL",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a team in an organization",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					orgFlag("Create the team in this org", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					createTeamCmd,
				),
			},
			{
				Name:  "list",
				Usage: "List teams in an organization",
				Flags: []cli.Flag{
					orgFlag("Use this organization.", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, teamsListCmd,
				),
			},
			{
				Name:      "members",
				Usage:     "List members of a particular team in and organization",
				ArgsUsage: "<team>",
				Flags: []cli.Flag{
					orgFlag("Use this organization.", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, teamMembersListCmd,
				),
			},
			{
				Name:      "add",
				ArgsUsage: "<username> <team>",
				Usage:     "Add user to a specified team in an organization you administer",
				Flags: []cli.Flag{
					stdOrgFlag,
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, teamsAddCmd,
				),
			},
			{
				Name:      "remove",
				Usage:     "Remove user from a specified team in an organization you administer",
				ArgsUsage: "<username> <team>",
				Flags: []cli.Flag{
					stdOrgFlag,
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, teamsRemoveCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, teams)
}

func teamsListCmd(ctx *cli.Context) error {

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
			return errs.NewExitError("Failed to select org.")
		}

		org = &orgs[idx]

	} else {
		// If org flag was used, identify the org supplied.
		org, err = client.Orgs.GetByName(c, orgName)
		if org == nil {
			return errs.NewExitError("org" + orgName + "not found.")
		}
	}

	var getMemberships, display sync.WaitGroup
	getMemberships.Add(1)
	display.Add(2)

	var teams []envelope.Team
	var session *api.Session
	var sErr, tErr error

	memberOf := make(map[identity.ID]bool)

	go func() {
		session, sErr = client.Session.Who(c)
		getMemberships.Done()
	}()

	go func() {
		teams, tErr = client.Teams.GetByOrg(c, org.ID)
		display.Done()
	}()

	go func() {
		getMemberships.Wait()
		var memberships []envelope.Membership
		if sErr == nil {
			memberships, sErr = client.Memberships.List(c, org.ID, nil, session.ID())
		}

		for _, m := range memberships {
			memberOf[*m.Body.TeamID] = true
		}
		display.Done()
	}()

	display.Wait()
	if sErr != nil || tErr != nil {
		return errs.FilterErrors(
			sErr,
			tErr,
			errs.NewExitError("Error fetching teams list"),
		)
	}

	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\t%s\t%s\n", ui.Bold("Team"), ui.Bold("Type"))
	for _, t := range teams {
		if isMachineTeam(t.Body) {
			continue
		}

		isMember := ""
		displayTeamType := ""

		switch teamType := t.Body.TeamType; teamType {
		case primitive.SystemTeamType:
			displayTeamType = "system"
		case primitive.MachineTeamType:
			displayTeamType = "machine"
		case primitive.UserTeamType:
			displayTeamType = "user"
		}

		if _, ok := memberOf[*t.ID]; ok {
			isMember = ui.Color(ansiterm.DarkGray, "*")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", isMember, t.Body.Name, displayTeamType)
	}

	w.Flush()

	count := strconv.Itoa(len(teams))
	var countStr string
	if len(teams) == 1 {
		countStr = "org " + org.Body.Name + " has (" + count + ") team."
	} else {
		countStr = "org " + org.Body.Name + " has (" + count + ") teams."
	}

	fmt.Println("")
	fmt.Println(countStr)
	fmt.Println("")
	return nil
}

func teamMembersListCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) < 1 || args[0] == "" {
		return errs.NewUsageExitError("Missing team name", ctx)
	}
	teamName := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	var getMembers sync.WaitGroup
	getMembers.Add(2)

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
			return errs.NewExitError("Failed to select org.")
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

	var team envelope.Team
	var teams []envelope.Team
	var memberships []envelope.Membership
	var tErr, mErr, sErr error
	go func() {
		// Retrieve the team by name supplied
		teams, tErr = client.Teams.GetByName(c, org.ID, teamName)
		if len(teams) != 1 {
			tErr = errs.NewExitError("Team not found.")
			getMembers.Done()
			return
		}
		team = teams[0]

		// Hide machine teams from the teams list; as we use them to represent
		// machine roles in the system.
		if isMachineTeam(team.Body) {
			tErr = errs.NewExitError("Team not found.")
			getMembers.Done()
			return
		}

		// Pull all memberships for supplied org/team
		memberships, mErr = client.Memberships.List(c, org.ID, team.ID, nil)
		getMembers.Done()
	}()

	var session *api.Session
	go func() {
		// Who am I
		session, sErr = client.Session.Who(c)
		getMembers.Done()
	}()

	getMembers.Wait()
	if mErr != nil || tErr != nil || sErr != nil {
		return errs.FilterErrors(mErr, tErr, sErr)
	}

	if len(memberships) == 0 {
		fmt.Printf("%s has no members\n", team.Body.Name)
		return nil
	}

	membershipUserIDs := make(map[identity.ID]bool)
	for _, membership := range memberships {
		membershipUserIDs[*membership.Body.OwnerID] = true
	}

	var profileIDs []identity.ID
	for id := range membershipUserIDs {
		profileIDs = append(profileIDs, id)
	}

	profiles, err := client.Profiles.ListByID(c, profileIDs)
	if err != nil {
		return err
	}
	if profiles == nil {
		return errs.NewExitError("User not found.")
	}

	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\t%s\t%s\n", ui.Bold("Name"), ui.Bold("Username"))
	for _, profile := range profiles {
		me := ""
		if session.Username() == profile.Body.Username {
			me = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", me, profile.Body.Name, ui.Color(ansiterm.DarkGray, profile.Body.Username))
	}

	w.Flush()

	count := strconv.Itoa(len(memberships))
	var countStr string
	if len(memberships) == 1 {
		countStr = "team " + team.Body.Name + " has (" + count + ") member."
	} else {
		countStr = "team " + team.Body.Name + " has (" + count + ") members."
	}

	fmt.Println("")
	fmt.Println(countStr)
	fmt.Println("")

	return nil
}

const teamCreateFailed = "Could not create team."

func createTeamCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewErrorExitError(teamCreateFailed, err)
	}

	args := ctx.Args()
	teamName := ""
	if len(args) > 0 {
		teamName = args[0]
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Ask the user which org they want to use
	org, oName, newOrg, err := SelectCreateOrg(c, client, ctx.String("org"))
	if err != nil {
		return handleSelectError(err, "Org selection failed")
	}
	if org == nil && !newOrg {
		fmt.Println("")
		return errs.NewExitError("Org not found.")
	}
	if newOrg && oName == "" {
		fmt.Println("")
		return errs.NewExitError("Invalid org name")
	}

	var orgID *identity.ID
	if org != nil {
		orgID = org.ID
	}

	label := "Team name"
	autoAccept := teamName != ""
	teamName, err = NamePrompt(&label, teamName, autoAccept)
	if err != nil {
		return handleSelectError(err, teamCreateFailed)
	}

	// Create the org now if needed
	if org == nil && newOrg {
		org, err = createOrgByName(c, ctx, client, oName)
		if err != nil {
			fmt.Println("")
			return err
		}
		orgID = org.ID
	}

	// Create our new team
	fmt.Println("")
	_, err = client.Teams.Create(c, orgID, teamName, primitive.UserTeamType)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Team already exists")
		}
		return errs.NewErrorExitError(teamCreateFailed, err)
	}

	fmt.Printf("Team %s created.\n", teamName)

	hints.Display(hints.Allow, hints.Deny)
	return nil
}

const teamRemoveFailed = "Failed to remove team member."

func teamsRemoveCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 2 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}
	if len(args) < 2 {
		return errs.NewUsageExitError("Too few arguments", ctx)
	}

	username := args[0]
	teamName := args[1]

	if username == "" {
		return errs.NewUsageExitError("Invalid username", ctx)
	}
	if teamName == "" {
		return errs.NewUsageExitError("Invalid team naem", ctx)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	var wait sync.WaitGroup
	wait.Add(2)

	var uErr, oErr, tErr error
	var org *envelope.Org
	var team envelope.Team
	var user *apitypes.Profile

	go func() {
		// Identify the org supplied
		result, err := client.Orgs.GetByName(c, ctx.String("org"))
		if result == nil || err != nil {
			oErr = errs.NewExitError("Org not found.")
			wait.Done()
			return
		}
		org = result

		// Retrieve the team by name supplied
		results, err := client.Teams.GetByName(c, org.ID, teamName)
		if len(results) != 1 || err != nil {
			tErr = errs.NewExitError("Team not found.")
		} else {
			team = results[0]
		}
		wait.Done()
	}()

	go func() {
		// Retrieve the user by name supplied
		result, err := client.Profiles.ListByName(c, username)
		if result == nil || err != nil {
			uErr = errs.NewExitError("User not found.")
		} else {
			user = result
		}
		wait.Done()
	}()

	wait.Wait()
	if uErr != nil || oErr != nil || tErr != nil {
		return cli.NewMultiError(
			oErr,
			uErr,
			tErr,
		)
	}

	// Lookup their membership row
	memberships, mErr := client.Memberships.List(c, org.ID, team.ID, user.ID)
	if mErr != nil || len(memberships) < 1 {
		return errs.NewExitError("Memberships not found.")
	}

	err = client.Memberships.Delete(c, memberships[0].ID)
	if err != nil {
		msg := teamRemoveFailed
		if strings.Contains(err.Error(), "member of the") {
			msg = "Must be a member of the admin team to remove members"
		}
		if strings.Contains(err.Error(), "cannot remove") {
			msg = "Cannot remove members from the member team"
		}
		return errs.NewExitError(msg)
	}

	fmt.Println(username + " has been removed from " + teamName + " team")
	return nil
}

const teamAddFailed = "Failed to add team member, please try again"

func teamsAddCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 2 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}
	if len(args) < 2 {
		return errs.NewUsageExitError("Too few arguments", ctx)
	}

	username := args[0]
	teamName := args[1]

	if username == "" {
		return errs.NewUsageExitError("Invalid username", ctx)
	}
	if teamName == "" {
		return errs.NewUsageExitError("Invalid team name", ctx)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	var wait sync.WaitGroup
	wait.Add(2)

	var uErr, oErr, tErr error
	var org *envelope.Org
	var team envelope.Team
	var user *apitypes.Profile

	go func() {
		// Identify the org supplied
		result, err := client.Orgs.GetByName(c, ctx.String("org"))
		if result == nil || err != nil {
			oErr = errs.NewExitError("Org not found.")
			wait.Done()
			return
		}
		org = result

		// Retrieve the team by name supplied
		results, err := client.Teams.GetByName(c, org.ID, teamName)
		if len(results) != 1 || err != nil {
			tErr = errs.NewExitError("Team not found.")
			wait.Done()
			return
		}
		team = results[0]
		wait.Done()
	}()

	go func() {
		// Retrieve the user by name supplied
		result, err := client.Profiles.ListByName(c, username)
		if result == nil || err != nil {
			uErr = errs.NewExitError("User not found.")
		} else {
			user = result
		}
		wait.Done()
	}()

	wait.Wait()
	if uErr != nil || oErr != nil || tErr != nil {
		return cli.NewMultiError(
			oErr,
			uErr,
			tErr,
		)
	}

	err = client.Memberships.Create(c, user.ID, org.ID, team.ID)
	if err != nil {
		msg := teamAddFailed
		if strings.Contains(err.Error(), "member of the") {
			msg = "Must be a member of the admin team to add members."
		}
		if strings.Contains(err.Error(), "resource exists") {
			msg = username + " is already a member of the " + teamName + " team."
		}
		if strings.Contains(err.Error(), "to the members team") {
			msg = username + " cannot be added to the " + teamName + " team."
		}
		return errs.NewExitError(msg)
	}

	fmt.Println(username + " has been added to the " + teamName + " team.")
	return nil
}

// isMachineTeam returns whether or not the given team represents a machine
// role (which uses the Team primitive)
func isMachineTeam(team *primitive.Team) bool {
	return team.TeamType == primitive.MachineTeamType || (team.TeamType == primitive.SystemTeamType && team.Name == primitive.MachineTeamName)
}
