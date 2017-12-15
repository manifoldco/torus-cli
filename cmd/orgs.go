package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	//"text/tabwriter"

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
	orgs := cli.Command{
		Name:     "orgs",
		Usage:    "View and create organizations",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new organization",
				ArgsUsage: "<name>",
				Action:    chain(ensureDaemon, ensureSession, orgsCreate),
			},
			{
				Name:   "list",
				Usage:  "List organizations associated with your account",
				Action: chain(ensureDaemon, ensureSession, orgsListCmd),
			},
			{
				Name:      "remove",
				Usage:     "Remove a user from an org",
				ArgsUsage: "<username>",
				Flags: []cli.Flag{
					orgFlag("org to remove the user from", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, orgsRemove,
				),
			},
			{
				Name:  "members",
				Usage: "List all members in an org",
				Flags: []cli.Flag{
					orgFlag("Use this organization.", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					orgsMembersListCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, orgs)
}

const orgCreateFailed = "Org creation failed."

func orgsCreate(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}

	var name string
	var err error

	if len(args) == 1 {
		name = args[0]
	}

	label := "Org name"
	autoAccept := name != ""
	name, err = NamePrompt(&label, name, autoAccept)
	if err != nil {
		return handleSelectError(err, orgCreateFailed)
	}

	c := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewErrorExitError(orgCreateFailed, err)
	}

	client := api.NewClient(cfg)

	_, err = createOrgByName(c, ctx, client, name)
	if err != nil {
		return err
	}

	hints.Display(hints.InvitesSend, hints.Projects, hints.Link)
	return nil
}

func createOrgByName(c context.Context, ctx *cli.Context, client *api.Client, name string) (*envelope.Org, error) {
	org, err := client.Orgs.Create(c, name)
	if err != nil {
		return nil, errs.NewErrorExitError(orgCreateFailed, err)
	}

	err = generateKeypairsForOrg(c, ctx, client, org.ID, false)
	if err != nil {
		msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
		return nil, errs.NewExitError(msg)
	}

	fmt.Println("Org " + org.Body.Name + " created.")
	return org, nil
}

func orgsListCmd(ctx *cli.Context) error {
	orgs, session, err := orgsList()
	if err != nil {
		return err
	}

	withoutPersonal := orgs

	fmt.Println("")
	fmt.Println("  " + ui.Bold("Name"))
	if session.Type() == apitypes.UserSession {
		for i, o := range orgs {
			if o.Body.Name == session.Username() {
				fmt.Printf("  %s %s\n", o.Body.Name, "(" + ui.Color(ansiterm.DarkGray, "personal") + ")")
				withoutPersonal = append(orgs[:i], orgs[i+1:]...)
			}
		}
	}

	for _, o := range withoutPersonal {
		fmt.Printf("  %s\n", o.Body.Name)
	}

	hints.Display(hints.PersonalOrg)

	return nil
}

func orgsList() ([]envelope.Org, *api.Session, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	client := api.NewClient(cfg)

	var wg sync.WaitGroup
	wg.Add(2)

	var orgs []envelope.Org
	var session *api.Session
	var oErr, sErr error

	go func() {
		orgs, oErr = client.Orgs.List(context.Background())
		wg.Done()
	}()

	go func() {
		session, sErr = client.Session.Who(context.Background())
		wg.Done()
	}()

	wg.Wait()
	if oErr != nil || sErr != nil {
		return nil, nil, errs.NewExitError("Error fetching orgs list")
	}

	return orgs, session, nil
}

func orgsRemove(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) < 1 || args[0] == "" {
		return errs.NewUsageExitError("Missing username", ctx)
	}
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}
	username := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	const userNotFound = "User not found."
	const orgsRemoveFailed = "Could not remove user from the org."

	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(orgsRemoveFailed, err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	profile, err := client.Profiles.ListByName(c, username)
	if apitypes.IsNotFoundError(err) {
		return errs.NewExitError(userNotFound)
	}
	if err != nil {
		return errs.NewErrorExitError(orgsRemoveFailed, err)
	}
	if profile == nil {
		return errs.NewExitError(userNotFound)
	}

	err = client.Orgs.RemoveMember(c, *org.ID, *profile.ID)
	if apitypes.IsNotFoundError(err) {
		fmt.Println("User is not a member of the org.")
		return nil
	}
	if err != nil {
		return errs.NewErrorExitError(orgsRemoveFailed, err)
	}

	fmt.Println("User has been removed from the org.")
	return nil
}

func orgsMembersListCmd(ctx *cli.Context) error {

	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError("Failed to load config.")
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

	// Retrieve teams, memberships and the current session concurrently
	var teams []envelope.Team
	var memberships []envelope.Membership
	var session *api.Session
	var tErr, mErr, sErr error

	var getMembersTeamsSession sync.WaitGroup
	getMembersTeamsSession.Add(3)

	go func() {
		// Retrieve list of teams in org
		teams, tErr = client.Teams.List(c, org.ID, "", primitive.AnyTeamType)
		if tErr != nil {
			tErr = errs.NewExitError("Could not retrieve list of teams.")
			getMembersTeamsSession.Done()
			return
		}

		getMembersTeamsSession.Done()
	}()

	go func() {
		// Retrieve list of memberships in org
		memberships, mErr = client.Memberships.List(c, org.ID, nil, nil)
		if mErr != nil {
			tErr = errs.NewExitError("Could not retrieve list of memberships.")
			getMembersTeamsSession.Done()
			return
		}
		getMembersTeamsSession.Done()
	}()

	go func() {
		// Get current session - Who am I
		session, err = client.Session.Who(c)
		if sErr != nil {
			sErr = errs.NewExitError("Failed to get current session.")
			getMembersTeamsSession.Done()
			return
		}
		getMembersTeamsSession.Done()
	}()

	getMembersTeamsSession.Wait()
	if tErr != nil || mErr != nil || sErr != nil {
		return cli.NewMultiError(
			tErr,
			mErr,
			sErr,
		)
	}

	if len(memberships) == 0 {
		return errs.NewExitError(org.Body.Name + " has no members.")
	}

	// Map team IDs to Team objects
	teamsIdx := make(map[identity.ID]envelope.Team)
	for _, team := range teams {
		teamsIdx[*team.ID] = team
	}

	userTeamIdx := make(map[identity.ID][]identity.ID) // Mapping from user ID to team IDs
	membershipUserIDs := make(map[identity.ID]bool)    // Set of unique user IDs

	// Create:
	//	- Set of unqiue user IDs in membershipUserIDs
	// 	- Mapping from user IDs to team IDs in userTeamIdx (1:m mapping)
	for _, membership := range memberships {
		// Skip memberships not associated with teams within org
		team, ok := teamsIdx[*membership.Body.TeamID]
		if !ok {
			panic("Attempted to access membership with no associated team.")
		}

		// Skip memberships associated with machine team
		if isMachineTeam(team.Body) {
			continue
		}

		// Add to set of unique user IDs
		membershipUserIDs[*membership.Body.OwnerID] = true

		// For each new user ID, create mapping to teams that the user is in
		_, ok = userTeamIdx[*membership.Body.OwnerID]
		if !ok {
			// For new user IDs (not yet in userTeamIdx) create an empty list of team IDs
			userTeamIdx[*membership.Body.OwnerID] = []identity.ID{}
		}
		// Add current membership's team ID to user's list
		userTeamIdx[*membership.Body.OwnerID] = append(userTeamIdx[*membership.Body.OwnerID], *membership.Body.TeamID)
	}

	// Create unique list of user IDs
	var userIDs []identity.ID
	for id := range membershipUserIDs {
		userIDs = append(userIDs, id)
	}

	users, err := client.Profiles.ListByID(c, userIDs)
	if err != nil {
		return errs.NewExitError("Could not list teams.")
	}
	if users == nil {
		return errs.NewExitError("User not found.")
	}

	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 2, 0, 3, ' ', 0)
	//fmt.Fprintf(w, " \t%s\t%s\t%s\n", ui.Bold("Username"), ui.Bold("Name"), ui.Bold("Team"))
	fmt.Fprintf(w, " \t%s\t%s\t%s\n", ui.Bold("Name"), ui.Bold("Username"), ui.Bold("Team"))
	for _, user := range users {
		me := ""
		if session.Username() == user.Body.Username {
			me = ui.Color(ansiterm.DarkGray, "*")
		}

		// Sort teams by precedence
		userTeams := []envelope.Team{}
		for _, teamID := range userTeamIdx[*user.ID] {
			userTeams = append(userTeams, teamsIdx[teamID])
		}

		sort.Sort(ByTeamPrecedence(userTeams))

		// Create string containing all team names associated with each user
		teamString := ""
		for _, t := range userTeams {
			teamString += t.Body.Name + ", "
		}
		// Remove trailing comma and space character
		teamString = teamString[:len(teamString)-2]
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", me, user.Body.Name, ui.Color(ansiterm.DarkGray, user.Body.Username), teamString)
	}

	w.Flush()

	count := strconv.Itoa(len(userIDs))
	var countStr string
	if len(userIDs) == 1 {
		countStr = "org " + org.Body.Name + " has (" + count + ") member."
	} else {
		countStr = "org " + org.Body.Name + " has (" + count + ") members."
	}

	fmt.Println("")
	fmt.Println(countStr)
	fmt.Println("")
	return nil
}

func getOrg(ctx context.Context, client *api.Client, name string) (*envelope.Org, error) {
	org, err := client.Orgs.GetByName(ctx, name)
	if err != nil {
		return nil, errs.NewErrorExitError("Unable to lookup org.", err)
	}
	if org == nil {
		return nil, errs.NewExitError("Org not found.")
	}

	return org, nil
}
