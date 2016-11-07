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
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
)

func init() {
	policies := cli.Command{
		Name:     "policies",
		Usage:    "View and manipulate access control list policies",
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
				Usage:     "Display the contents of an ACL policy",
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
				Usage:     "Detach a policy from a team or machine role, does not delete the policy",
				ArgsUsage: "<name> <team|role>",
				Flags: []cli.Flag{
					orgFlag("org to detach policy from", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, detachPolicies,
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
	var org *api.OrgResult
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(policyDetachFailed, err)
	}
	if org == nil {
		return errs.NewExitError("Org not found")
	}

	var waitPolicy sync.WaitGroup
	waitPolicy.Add(2)

	var team *api.TeamResult
	var policy *api.PoliciesResult
	var pErr, tErr error

	go func() {
		var policies []api.PoliciesResult
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
	var org *api.OrgResult
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

	var policies []api.PoliciesResult
	var pErr error
	go func() {
		policies, pErr = client.Policies.List(c, org.ID, "")
		getAttachments.Done()
	}()

	var attachments []api.PolicyAttachmentResult
	var aErr error
	go func() {
		attachments, aErr = client.Policies.AttachmentsList(c, org.ID, nil, nil)
		getAttachments.Done()
	}()

	var teams []api.TeamResult
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

	teamsByID := make(map[identity.ID]api.TeamResult)
	policiesByName := make(map[string]api.PoliciesResult)
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
