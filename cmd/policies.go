package cmd

import (
	"context"
	"fmt"
	"log"
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

var permissionString = map[bool]string{
	true:  "yes",
	false: "no",
}

func testPolicies(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	action, userName, path, err := parseArgs(ctx)
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(policyTestFailed, err)
	} else if org == nil {
		return errs.NewExitError("Org not found")
	}

	// Get the Teams (to which the user is a member), Policies and
	// PolicyAttachments concurrently.
	wg := &sync.WaitGroup{}
	wg.Add(2)
	var teamsErr error
	var teams []envelope.Team
	go func() {
		defer wg.Done()
		teams, teamsErr = getTeamsForUser(c, client, userName, org.ID)
	}()

	var policiesErr error
	var policies []envelope.Policy
	var attachments []envelope.PolicyAttachment
	go func() {
		defer wg.Done()
		policies, attachments, policiesErr = getPoliciesAndAttachments(c, client, org.ID)
	}()
	wg.Wait()

	if teamsErr != nil || policiesErr != nil {
		return cli.NewMultiError(teamsErr, policiesErr,
			errs.NewExitError(policyTestFailed))
	}

	//whittle the policies down to those that are relevant.
	predicate := AllPredicate(
		policyAttachedToTeamsPredicate(teams, attachments),
		policyTouchesPathPredicate(path),
		policyImplementsActionPredicate(*action),
	)
	policies = filterPolicies(policies, predicate)
	allowed := policiesAllowAccess(policies)

	fmt.Println(permissionString[allowed])

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
	path, _, err := parseRawPath(rawPath)
	if err != nil {
		return nil, nil, nil, errs.NewErrorExitError(policyTestFailed, err)
	}
	return &action, &userName, path, nil
}

// getTeamsForUser fetches all teams to which the user belongs. To do this it
// must first get all teams for the org and membership information for each
// team, then filter (include) teams if their membership includes the user.
// Since fetching user, teams, and memberships are expensive, do them in
// parallel as much as possible.
func getTeamsForUser(c context.Context, client *api.Client, userName *string,
	orgID *identity.ID) ([]envelope.Team, error) {
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
	filterTeamsByUser(c, client, teams, user, teamChan, orgID)

	var result []envelope.Team
	for t := range teamChan {
		result = append(result, t)
	}
	return result, nil
}

// filterTeamsByUser, given a set of teams and a user, filters (excludes) teams
// for which the user is not a member. This requires getting the membership
// information for each team, which can be expensive, but can be done in
// parallel.
func filterTeamsByUser(c context.Context, client *api.Client, teams []envelope.Team,
	user *apitypes.Profile, teamChan chan envelope.Team, orgID *identity.ID) {

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
					log.Printf("Failed to list memberships for team %v: %v. Skipping...\n", team.Body.Name, err)
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

// getPoliciesAndAttachments fetches all policies and policy-attachments for the
// org. The two steps are independent so can be (and are) done in parallel.
func getPoliciesAndAttachments(c context.Context, client *api.Client,
	orgID *identity.ID) ([]envelope.Policy, []envelope.PolicyAttachment, error) {

	wg := &sync.WaitGroup{}
	wg.Add(2)

	var pErr error
	var policies []envelope.Policy
	go func() {
		defer wg.Done()
		policies, pErr = client.Policies.List(c, orgID, "")
	}()

	var aErr error
	var attachments []envelope.PolicyAttachment
	go func() {
		defer wg.Done()
		attachments, aErr = client.Policies.AttachmentsList(c, orgID, nil, nil)
	}()
	wg.Wait()

	if aErr != nil || pErr != nil {
		return nil, nil, cli.NewMultiError(pErr, aErr,
			errs.NewExitError(policyTestFailed))
	}
	return policies, attachments, nil
}

// PolicyPredicate is a signature for a function that tests if a Policy
// satisfies some condition
type PolicyPredicate func(envelope.Policy) bool

// AllPredicate is a compound Predicate that tests if a policy satisfies all
// predicates
func AllPredicate(predicates ...PolicyPredicate) PolicyPredicate {
	return func(policy envelope.Policy) bool {
		for _, pred := range predicates {
			if !pred(policy) {
				return false
			}
		}
		return true
	}
}

func filterPolicies(policies []envelope.Policy, predicate PolicyPredicate) []envelope.Policy {
	var result []envelope.Policy
	for _, pol := range policies {
		if predicate(pol) {
			result = append(result, pol)
		}
	}
	return result
}

// policyAttachedToTeamsPredicate cretes a Predicate to filter (exclude)
// policies that are not attached to any of the specified teams.
func policyAttachedToTeamsPredicate(teams []envelope.Team,
	attachments []envelope.PolicyAttachment) PolicyPredicate {

	teamsByID := make(map[identity.ID]envelope.Team)
	for _, t := range teams {
		teamsByID[*t.ID] = t
	}

	attachmentByPolicyID := make(map[identity.ID]envelope.PolicyAttachment, 0)
	for _, a := range attachments {
		attachmentByPolicyID[*a.Body.PolicyID] = a
	}

	return func(policy envelope.Policy) bool {
		if a, ok := attachmentByPolicyID[*policy.ID]; ok {
			if _, ok := teamsByID[*a.Body.OwnerID]; ok {
				return true
			}
		}
		return false
	}
}

// policyTouchesPathPredicate creates a predicate to filter (exclude) policies
// that do not apply to the specified path. A policy applies to a path if any of
// its resources (paths) are equivalent to the specified path. The secret
// portion of the path is ignored.
func policyTouchesPathPredicate(pathExp *pathexp.PathExp) PolicyPredicate {
	return func(policy envelope.Policy) bool {
		for _, s := range policy.Body.Policy.Statements {
			rp, _, err := parseRawPath(s.Resource)
			if err != nil {
				log.Printf("Failed parse resource %v for policy %v: %v. Skipping...\n",
					s.Resource, policy.Body.Policy.Name, err)
				continue
			}
			// I'm pretty sure this is wrong...
			if rp.Equal(pathExp) {
				return true
			}
		}
		return false
	}
}

// policyImplementsActionPredicate creates a predicate to filter (include)
// policies that implement the specified policy-action ([crudl]) in any one of
// their statements.
func policyImplementsActionPredicate(action primitive.PolicyAction) PolicyPredicate {
	return func(policy envelope.Policy) bool {
		for _, s := range policy.Body.Policy.Statements {
			if s.Action&action > 0 {
				return true
			}
		}
		return false
	}
}

// policiesAllowAccess determines if a set of polices "allow access" (i.e. allow
// a specific action on a specific path). Returns true if at least one policy
// allows access AND exactly zero policies deny access. Return false otherwise.
// This method only make sense if the policies have been previously reduced to
// those referring to the same resource and action.
func policiesAllowAccess(policies []envelope.Policy) bool {
	allow := false
	for _, p := range policies {
		for _, s := range p.Body.Policy.Statements {
			if s.Effect == primitive.PolicyEffectDeny {
				return false
			}
			allow = true
		}
	}
	return allow

}
