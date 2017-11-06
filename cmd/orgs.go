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
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/promptui"
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
					setUserEnv, checkRequiredFlags, orgsMembersListCmd,
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

	if session.Type() == apitypes.UserSession {
		for i, o := range orgs {
			if o.Body.Name == session.Username() {
				fmt.Printf("  %s [personal]\n", o.Body.Name)
				withoutPersonal = append(orgs[:i], orgs[i+1:]...)
			}
		}
	}

	for _, o := range withoutPersonal {
		fmt.Printf("  %s\n", o.Body.Name)
	}

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

	// Use "member" team name to bypass specific teams
	teamName := "member"

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	var org *envelope.Org
	var orgs []envelope.Org
	var team envelope.Team
	var teams []envelope.Team
	var memberships []envelope.Membership
	var oErr, tErr, mErr, sErr error

	// Retrieve the org name supplied via the --org flag.
	// This flag is optional. If none was supplied, then
	// orgFlagArgument will be set to "". In this case,
	// prompt the user to select an org.
	orgFlagArgument := ctx.String("org")

	if orgFlagArgument == "" {
		// Retrieve list of available orgs
		orgs, oErr = client.Orgs.List(c)
		if oErr != nil {
			return oErr
		}

		//idx, _, oErr := SelectOrgPrompt(orgs)
		idx, _, oErr := SelectExistingOrgPrompt(orgs)
		if oErr != nil {
			return oErr
		}

		if idx == promptui.SelectedAdd {
			// COME BACK TO THIS
			// Error condition, unsure how to handle
		}

		org = &orgs[idx]
	} else {
		// If org flag was used, identify the org supplied.
		org, oErr = client.Orgs.GetByName(c, orgFlagArgument)
		if org == nil {
			return oErr
		}
	}

	// Retrieve the team by name supplied
	teams, tErr = client.Teams.GetByName(c, org.ID, teamName)
	if len(teams) != 1 {
		return tErr
	}
	team = teams[0]

	// Hide machine teams from the teams list; as we use them to represent
	// machine roles in the system.
	if isMachineTeam(team.Body) {
		return tErr
	}

	// Pull all memberships for supplied org/team
	memberships, mErr = client.Memberships.List(c, org.ID, team.ID, nil)

	var session *api.Session
	// Who am I
	session, sErr = client.Session.Who(c)

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
	for _, profile := range profiles {
		me := ""
		if session.Username() == profile.Body.Username {
			me = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", me, profile.Body.Name, profile.Body.Username)
	}

	w.Flush()
	fmt.Println("\n  (*) you")
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
