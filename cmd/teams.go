package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

func init() {
	teams := cli.Command{
		Name:     "teams",
		Usage:    "View and manipulate teams within an organization",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:      "members",
				Usage:     "List members of a particular team in and organization",
				ArgsUsage: "<team>",
				Flags: []cli.Flag{
					stdOrgFlag,
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, teamMembersListCmd,
				),
			},
			{
				Name:  "list",
				Usage: "List teams in an organization",
				Flags: []cli.Flag{
					stdOrgFlag,
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, teamsListCmd,
				),
			},
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
		},
	}
	Cmds = append(Cmds, teams)
}

func teamsListCmd(ctx *cli.Context) error {
	orgName := ctx.String("org")

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	var getMemberships, display sync.WaitGroup
	getMemberships.Add(2)
	display.Add(2)

	var teams []api.TeamResult
	var org *api.OrgResult
	var self *api.UserResult
	var oErr, sErr, tErr error

	memberOf := make(map[identity.ID]bool)

	c := context.Background()

	go func() {
		self, sErr = client.Users.Self(c)
		getMemberships.Done()
	}()

	go func() {
		org, oErr = client.Orgs.GetByName(c, orgName)
		getMemberships.Done()

		if org == nil {
			oErr = errs.NewExitError("Org not found.")
			display.Done()
			return
		}

		teams, tErr = client.Teams.GetByOrg(c, org.ID)
		display.Done()
	}()

	go func() {
		getMemberships.Wait()
		var memberships []api.MembershipResult
		if oErr == nil && sErr == nil {
			memberships, sErr = client.Memberships.List(c, org.ID, self.ID, nil)
		}

		for _, m := range memberships {
			memberOf[*m.Body.TeamID] = true
		}
		display.Done()
	}()

	display.Wait()
	if oErr != nil || sErr != nil || tErr != nil {
		return cli.NewMultiError(
			oErr,
			sErr,
			tErr,
			errs.NewExitError("Error fetching teams list"),
		)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	for _, t := range teams {
		isMember := ""
		teamType := ""
		if t.Body.TeamType == primitive.SystemTeam {
			teamType = "[system]"
		}

		if _, ok := memberOf[*t.ID]; ok {
			isMember = "*"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", isMember, t.Body.Name, teamType)
	}

	w.Flush()
	fmt.Println("\n  (*) member")
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

	var org *api.OrgResult
	var team api.TeamResult
	var teams []api.TeamResult
	var memberships []api.MembershipResult
	var oErr, tErr, mErr, sErr error
	go func() {
		// Identify the org supplied
		org, oErr = client.Orgs.GetByName(c, ctx.String("org"))
		if org == nil {
			oErr = errs.NewExitError("Org not found.")
			getMembers.Done()
			return
		}

		// Retrieve the team by name supplied
		teams, tErr = client.Teams.GetByName(c, org.ID, teamName)
		if len(teams) != 1 {
			tErr = errs.NewExitError("Team not found.")
			getMembers.Done()
			return
		}
		team = teams[0]

		// Pull all memberships for supplied org/team
		memberships, mErr = client.Memberships.List(c, org.ID, nil, team.ID)
		getMembers.Done()
	}()

	var self *api.UserResult
	go func() {
		// Who am I
		self, sErr = client.Users.Self(c)
		getMembers.Done()
	}()

	getMembers.Wait()
	if oErr != nil || mErr != nil || tErr != nil {
		return cli.NewMultiError(
			oErr,
			mErr,
			tErr,
			sErr,
		)
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

	count := strconv.Itoa(len(memberships))
	title := "members of the " + team.Body.Name + " team (" + count + ")"

	fmt.Println("")
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", utf8.RuneCountInString(title)))

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	for _, profile := range *profiles {
		me := ""
		if self.Body.Username == profile.Body.Username {
			me = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", me, profile.Body.Name, profile.Body.Username)
	}

	w.Flush()
	fmt.Println("\n  (*) you")
	return nil
}

const teamCreateFailed = "Could not create team. Please try again."

func createTeamCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(teamCreateFailed)
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
	_, err = client.Teams.Create(c, orgID, teamName, "")
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Team already exists")
		}
		return errs.NewErrorExitError(teamCreateFailed, err)
	}

	fmt.Printf("Team %s created.\n", teamName)
	return nil
}

const teamRemoveFailed = "Failed to remove team member, please try again"

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
	var org *api.OrgResult
	var team api.TeamResult
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
	memberships, mErr := client.Memberships.List(c, org.ID, user.ID, team.ID)
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
	var org *api.OrgResult
	var team api.TeamResult
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
