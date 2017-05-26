package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
)

func init() {
	policies := cli.Command{
		Name:     "policies",
		Usage:    "Manage which resources machines and users can access",
		Category: "ACCESS CONTROL",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List ACL policies for an organization",
				Flags: []cli.Flag{
					orgFlag("org to show policies for", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, listPolicies,
				),
			},
			{
				Name:      "view",
				Usage:     "Display the contents of a policy",
				ArgsUsage: "<policy>",
				Flags: []cli.Flag{
					orgFlag("org to show policies for", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, viewPolicyCmd,
				),
			},

			{
				Name:      "detach",
				Usage:     "Detach (but not delete) a policy from a team or role",
				ArgsUsage: "<name> <team|role>",
				Flags: []cli.Flag{
					orgFlag("org to detach policy from", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, detachPolicies,
				),
			},

			{
				Name:      "test",
				Usage:     "Test a user's access to a path",
				ArgsUsage: "<c|r|u|d|l> <username> <path>",
				Flags: []cli.Flag{
					orgFlag("org to test policy for", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, testPolicies,
				),
			},
		},
	}
	Cmds = append(Cmds, policies)
}

const policyDetachFailed = "Could not detach policy."

func detachPolicies(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	args := ctx.Args()
	if len(args) < 2 {
		return errs.NewUsageExitError("Too few arguments", ctx)

	} else if len(args) > 2 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}

	policyName := args[0]
	teamName := args[1]

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	var org *envelope.Org
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(policyDetachFailed, err)
	}
	if org == nil {
		return errs.NewExitError("Org not found")
	}

	var waitPolicy sync.WaitGroup
	waitPolicy.Add(2)

	var team *envelope.Team
	var policy *envelope.Policy
	var pErr, tErr error

	go func() {
		var policies []envelope.Policy
		policies, pErr = client.Policies.List(c, org.ID, "")
		for _, p := range policies {
			if p.Body.Policy.Name == policyName {
				policy = &p
				break
			}
		}
		waitPolicy.Done()
	}()

	go func() {
		teams, tErr := client.Teams.GetByName(c, org.ID, teamName)
		if len(teams) < 1 || tErr != nil {
			waitPolicy.Done()
			return
		}
		team = &teams[0]
		waitPolicy.Done()
	}()

	waitPolicy.Wait()
	if tErr != nil || pErr != nil {
		return cli.NewMultiError(
			tErr,
			pErr,
		)
	}
	if team == nil {
		return errs.NewExitError("Team " + teamName + " not found.")
	}
	if policy == nil {
		return errs.NewExitError("Policy " + policyName + " not found.")
	}

	attachments, err := client.Policies.AttachmentsList(c, org.ID, team.ID, policy.ID)
	if err != nil {
		return errs.NewErrorExitError(policyDetachFailed, err)
	}
	if len(attachments) < 1 {
		return errs.NewExitError(policyName + " policy is not currently attached to " + teamName)
	}

	err = client.Policies.Detach(c, attachments[0].ID)
	if err != nil {
		if strings.Contains(err.Error(), "system team") {
			return errs.NewExitError("Cannot delete system team attachment")
		}
		return errs.NewErrorExitError(policyDetachFailed, err)
	}

	fmt.Println("Policy " + policyName + " has been detached from team " + teamName)
	return nil
}

const policyListFailed = "Could not list policies."

func listPolicies(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	var org *envelope.Org
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(policyListFailed, err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	var getAttachments, display sync.WaitGroup
	getAttachments.Add(3)
	display.Add(1)

	var policies []envelope.Policy
	var pErr error
	go func() {
		policies, pErr = client.Policies.List(c, org.ID, "")
		getAttachments.Done()
	}()

	var attachments []envelope.PolicyAttachment
	var aErr error
	go func() {
		attachments, aErr = client.Policies.AttachmentsList(c, org.ID, nil, nil)
		getAttachments.Done()
	}()

	var teams []envelope.Team
	var tErr error
	go func() {
		teams, tErr = client.Teams.GetByOrg(c, org.ID)
		getAttachments.Done()
	}()

	if aErr != nil || pErr != nil || tErr != nil {
		return cli.NewMultiError(
			pErr,
			aErr,
			tErr,
			errs.NewExitError(policyListFailed),
		)
	}

	teamsByID := make(map[identity.ID]envelope.Team)
	policiesByName := make(map[string]envelope.Policy)
	attachedTeamsByPolicyID := make(map[identity.ID][]string)
	var sortedNames []string

	go func() {
		getAttachments.Wait()
		for _, t := range teams {
			teamsByID[*t.ID] = t
		}
		for _, p := range policies {
			policiesByName[p.Body.Policy.Name] = p
			sortedNames = append(sortedNames, p.Body.Policy.Name)
		}
		sort.Strings(sortedNames)
		for _, a := range attachments {
			ID := *a.Body.PolicyID
			attachedTeamsByPolicyID[ID] = append(attachedTeamsByPolicyID[ID], teamsByID[*a.Body.OwnerID].Body.Name)
		}
		display.Done()
	}()

	display.Wait()
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "POLICY NAME\tTYPE\tATTACHED TO")
	fmt.Fprintln(w, " \t \t ")
	for _, name := range sortedNames {
		teamNames := ""
		policy := policiesByName[name]
		policyID := *policy.ID
		if len(attachedTeamsByPolicyID[policyID]) > 0 {
			teamNames = strings.Join(attachedTeamsByPolicyID[policyID], ", ")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", policy.Body.Policy.Name, policy.Body.PolicyType, teamNames)
	}

	w.Flush()
	fmt.Println("")
	return nil
}

func viewPolicyCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 1 {
		msg := "policy name is required."
		if len(args) > 1 {
			msg = "Too many arguments provided."
		}
		return errs.NewUsageExitError(msg, ctx)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError("Unable to lookup org.", err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	policies, err := client.Policies.List(c, org.ID, args[0])
	if err != nil {
		return errs.NewExitError("Unable to list policies.")
	}

	if len(policies) < 1 {
		return errs.NewExitError("Policy '" + args[0] + "' not found.")
	}

	policy := policies[0]
	p := policy.Body.Policy

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)

	fmt.Fprintf(w, "Name:\t%s\n", p.Name)
	fmt.Fprintf(w, "Description:\t%s\n", p.Description)
	fmt.Fprintln(w, "")
	w.Flush()

	for _, stmt := range p.Statements {
		fmt.Fprintf(w, "%s\t%s\t%s\n", stmt.Effect.String(), stmt.Action.ShortString(), stmt.Resource)
	}
	w.Flush()

	return nil
}

const policyTestFailed = "Could not test policy."

func testPolicies(ctx *cli.Context) error {
	action, userName, path, err := parseArgs(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("User %s has access to %v %v: %v\n", *userName, action.String(), path, false)

	return nil
}

func parseArgs(ctx *cli.Context) (*primitive.PolicyAction, *string, *pathexp.PathExp, error) {
	args := ctx.Args()
	if len(args) < 3 {
		return nil, nil, nil, errs.NewUsageExitError("Too few arguments", ctx)
	} else if len(args) > 3 {
		return nil, nil, nil, errs.NewUsageExitError("Too many arguments", ctx)
	}

	rawAction := args[0]
	userName := args[1]
	rawPath := args[2]

	//Validate action
	action, err := parseAction(rawAction)
	if err != nil {
		return nil, nil, nil, errs.NewErrorExitError(policyTestFailed, err)
	}

	// Validate path
	path, err := parseRawPath(rawPath)
	if err != nil {
		return nil, nil, nil, errs.NewErrorExitError(policyTestFailed, err)
	}
	return &action, &userName, path, nil
}

// Parse and validate a raw path, possibly including a secret. The secret
// portion is discarded.
func parseRawPath(rawPath string) (*pathexp.PathExp, error) {
	idx := strings.LastIndex(rawPath, "/")
	if idx == -1 {
		msg := "resource path format is incorrect."
		return nil, errs.NewExitError(msg)
	}
	name := rawPath[idx+1:]
	path := rawPath[:idx]

	if name == "**" {
		path = rawPath
		name = "*"
	}

	// Ensure that the secret name is valid
	if !pathexp.ValidSecret(name) {
		return nil, errs.NewExitError("Invalid secret name")
	}

	pe, err := pathexp.Parse(path)
	if err != nil {
		return nil, errs.NewErrorExitError("Invalid path expression", err)
	}
	return pe, nil
}

// Fetch all teams to which the user belongs. To do this we must first get all
// teams for the org and membership information for each team, then filter
// (include) teams if their membership includes the user.
// Since fetching user, teams, and memberships are expensive, do them in
// parallel as much as possible.
func getTeamsForUser(client *api.Client, userName *string, orgID *identity.ID) ([]envelope.Team, error) {
	c := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	var userErr error
	var user *apitypes.Profile
	go func() {
		defer wg.Done()
		user, userErr = client.Profiles.ListByName(c, *userName)
		if user.Body == nil || user.ID == nil {
			userErr = errs.NewExitError("Could not find user " + *userName)
		}
	}()

	var teamErr error
	var teams []envelope.Team
	go func() {
		defer wg.Done()
		teams, teamErr = client.Teams.GetByOrg(c, orgID)
		if teamErr != nil {
			return
		}

		if teams == nil || len(teams) == 0 {
			teamErr = errs.NewExitError("No teams found for organisation.")
			return
		}
	}()
	wg.Wait()

	if userErr != nil || teamErr != nil {
		return nil, cli.NewMultiError(userErr, teamErr,
			errs.NewExitError(policyTestFailed),
		)
	}

	teamChan := make(chan envelope.Team, 1)
	filterTeamsByUser(client, teams, user, teamChan, orgID)

	result := make([]envelope.Team, 0)
	for t := range teamChan {
		result = append(result, t)
	}
	return result, nil
}

// Given a set of teams and a user, filter (exclude) the teams for which the
// user is not a member. This requires getting the membership information for
// each team, which can be expensive, but can be done in parallel.
func filterTeamsByUser(client *api.Client, teams []envelope.Team,
	user *apitypes.Profile, teamChan chan envelope.Team,
	orgID *identity.ID) {

	c := context.Background()
	wg := &sync.WaitGroup{}
	go func() {
		for _, t := range teams {
			if isMachineTeam(t.Body) {
				continue
			}

			wg.Add(1)
			go func(team envelope.Team) {
				defer wg.Done()
				members, err := client.Memberships.List(c, orgID, team.ID, nil)
				if err != nil {
					// TODO: Log this!!!
					return
				}
				for _, m := range members {
					if *user.ID == *m.Body.OwnerID {
						teamChan <- team
						return
					}
				}
			}(t)
		}
		wg.Wait()
		close(teamChan)
	}()
}
