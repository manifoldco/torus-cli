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

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
	"github.com/arigatomachine/cli/promptui"
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
					StdOrgFlag,
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, teamMembersListCmd,
				),
			},
			{
				Name:  "list",
				Usage: "List teams in an organization",
				Flags: []cli.Flag{
					StdOrgFlag,
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, teamsListCmd,
				),
			},
			{
				Name:      "create",
				Usage:     "Create a team in an organization",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					OrgFlag("Create the team in this org", false),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					createTeamCmd,
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
	var oErr, sErr error

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
			oErr = cli.NewExitError("Org not found", -1)
		}

		if oErr == nil {
			teams, oErr = client.Teams.GetByOrg(c, org.ID)
		}

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
	if oErr != nil || sErr != nil {
		return cli.NewMultiError(
			oErr,
			sErr,
			cli.NewExitError("Error fetching teams list", -1),
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
	usage := usageString(ctx)

	args := ctx.Args()
	if len(args) < 1 || args[0] == "" {
		text := "Missing team name\n\n"
		text += usage
		return cli.NewExitError(text, -1)
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
			oErr = cli.NewExitError("Org not found", -1)
		}

		// Retrieve the team by name supplied
		teams, tErr = client.Teams.GetByName(c, org.ID, teamName)
		if len(teams) != 1 {
			tErr = cli.NewExitError("Team not found", -1)
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

	count := strconv.Itoa(len(memberships))
	title := "members of the " + team.Body.Name + " team (" + count + ")"

	fmt.Println("")
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", utf8.RuneCountInString(title)))

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	for _, profile := range profiles {
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
		return cli.NewExitError(teamCreateFailed, -1)
	}

	args := ctx.Args()
	teamName := ""
	if len(args) > 0 {
		teamName = args[0]
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Ask the user which org they want to use
	org, oName, newOrg, err := SelectCreateOrg(client, c, ctx.String("org"))
	if err != nil {
		return handleSelectError(err, "Org selection failed")
	}
	if org == nil && !newOrg {
		fmt.Println("")
		return cli.NewExitError("Org not found", -1)
	}
	if newOrg && oName == "" {
		fmt.Println("")
		return cli.NewExitError("Invalid org name", -1)
	}

	var orgID *identity.ID
	if org != nil {
		orgID = org.ID
	}

	label := "Team name"
	if teamName == "" {
		teamName, err = NamePrompt(&label, "")
		if err != nil {
			return handleSelectError(err, teamCreateFailed)
		}
	} else {
		fmt.Println(promptui.SuccessfulValue(label, teamName))
	}

	// Create the org now if needed
	if org == nil && newOrg {
		org, err = createOrgByName(ctx, c, client, oName)
		if err != nil {
			fmt.Println("")
			return err
		}
		orgID = org.ID
	}

	// Create our new team
	fmt.Println("")
	err = client.Teams.Create(c, orgID, teamName)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return cli.NewExitError("Team already exists", -1)
		}
		fmt.Println(err)
		return cli.NewExitError(teamCreateFailed, -1)
	}

	fmt.Printf("Team %s created.\n", teamName)
	return nil
}
