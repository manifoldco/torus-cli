package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/ui"
	"github.com/manifoldco/torus-cli/validate"
)

func init() {
	policies := cli.Command{
		Name:     "policies",
		Usage:    "Manage which resources machines and users can access",
		Category: "ACCESS CONTROL",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List all policies for an organization",
				Flags: []cli.Flag{
					orgFlag("The org to show policies for", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, listPoliciesCmd,
				),
			},
			{
				Name:      "view",
				Usage:     "Display the contents of a policy",
				ArgsUsage: "<policy>",
				Flags: []cli.Flag{
					orgFlag("The org the policy belongs to", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, viewPolicyCmd,
				),
			},

			{
				Name:      "detach",
				Usage:     "Detach (but not delete) a policy from a team or machine role",
				ArgsUsage: "<name> <team|machine-role>",
				Flags: []cli.Flag{
					orgFlag("The org the team and policy belong to", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, detachPolicyCmd,
				),
			},

			{
				Name:      "attach",
				Usage:     "Attach a policy to a team or machine role",
				ArgsUsage: "<name> <team|machine-role>",
				Flags: []cli.Flag{
					orgFlag("The org the team and policy belong to", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, attachPolicyCmd,
				),
			},

			{
				Name:      "delete",
				Usage:     "Delete a policy from the organization",
				ArgsUsage: "<name>",
				Flags: []cli.Flag{
					stdAutoAcceptFlag,
					orgFlag("The org the policy belongs to", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, deletePolicyCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, policies)
}

const policyDetachFailed = "Could not detach policy."
const policyAttachFailed = "Could not attach policy."
const policyDeleteFailed = "Could not delete policy."

func attachPolicyCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	args := ctx.Args()
	if len(args) < 2 {
		return errs.NewUsageExitError("Not enough arguments provided", ctx)
	} else if len(args) > 2 {
		return errs.NewUsageExitError("Too many arguments provided", ctx)
	}

	policyName := args[0]
	teamName := args[1]

	if err := validate.PolicyName(policyName); err != nil {
		return errs.NewUsageExitError("Invalid policy name provided", ctx)
	}
	if err := validate.TeamName(teamName); err != nil {
		return errs.NewUsageExitError("Invalid team name provided", ctx)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, policy, team, err := getOrgPolicyAndTeam(c, client, ctx.String("org"), policyName, teamName)
	if err != nil {
		return err
	}

	err = client.Policies.Attach(c, org.ID, policy.ID, team.ID)
	if err != nil {
		return errs.NewErrorExitError(policyAttachFailed, err)
	}

	fmt.Printf("Policy %s has been attached to team %s!\n",
		policy.Body.Policy.Name,
		team.Body.Name,
	)
	return nil
}

func getOrgPolicyAndTeam(ctx context.Context, client *api.Client, orgName,
	policyName, teamName string) (*envelope.Org, *envelope.Policy, *envelope.Team, error) {

	org, err := client.Orgs.GetByName(ctx, orgName)
	if err != nil {
		return nil, nil, nil, errs.NewErrorExitError(policyDetachFailed, err)
	}
	if org == nil {
		return nil, nil, nil, errs.NewExitError("Org not found")
	}

	var waitPolicy sync.WaitGroup
	waitPolicy.Add(2)

	var team *envelope.Team
	var policy *envelope.Policy
	var pErr, tErr error

	go func() {
		var policies []envelope.Policy
		policies, pErr = client.Policies.List(ctx, org.ID, "")
		for _, p := range policies {
			if p.Body.Policy.Name == policyName {
				policy = &p
				break
			}
		}
		waitPolicy.Done()
	}()

	go func() {
		teams, tErr := client.Teams.GetByName(ctx, org.ID, teamName)
		if len(teams) < 1 || tErr != nil {
			waitPolicy.Done()
			return
		}
		team = &teams[0]
		waitPolicy.Done()
	}()

	waitPolicy.Wait()
	if tErr != nil || pErr != nil {
		return nil, nil, nil, errs.MultiError(
			tErr,
			pErr,
		)
	}
	if team == nil {
		return nil, nil, nil, errs.NewExitError("Team " + teamName + " not found.")
	}
	if policy == nil {
		return nil, nil, nil, errs.NewExitError("Policy " + policyName + " not found.")
	}

	return org, policy, team, nil
}

func detachPolicyCmd(ctx *cli.Context) error {
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

	if err := validate.PolicyName(policyName); err != nil {
		return errs.NewUsageExitError("Invalid policy name provided", ctx)
	}
	if err := validate.TeamName(teamName); err != nil {
		return errs.NewUsageExitError("Invalid team name provided", ctx)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	org, policy, team, err := getOrgPolicyAndTeam(c, client, ctx.String("org"), policyName, teamName)
	if err != nil {
		return err
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

func deletePolicyCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	args := ctx.Args()
	if len(args) != 1 {
		return errs.NewUsageExitError("A policy name must be provided", ctx)
	}

	policyName := args[0]
	if err := validate.PolicyName(policyName); err != nil {
		return errs.NewUsageExitError("Invalid policy name provided", ctx)
	}

	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(policyDeleteFailed, err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	policies, err := client.Policies.List(c, org.ID, policyName)
	if err != nil {
		return errs.NewErrorExitError(policyDeleteFailed, err)
	}

	if len(policies) == 0 {
		return errs.NewExitError("Policy not found.")
	}

	policy := policies[0]
	preamble := fmt.Sprintf("You are about to delete the %s policy and all "+
		"of it's attachments. This cannot be undone.", policyName)
	err = ConfirmDialogue(ctx, nil, &preamble, "", true) // Will error if user does not confirm
	if err != nil {
		return err
	}

	err = client.Policies.Delete(c, policy.ID)
	if err != nil {
		return errs.NewErrorExitError(policyDeleteFailed, err)
	}

	fmt.Printf("\nPolicy %s and all of it's attachments have been deleted.\n", policyName)
	return nil
}

func listPoliciesCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := getOrgWithPrompt(client, c, ctx.String("org"))
	if err != nil {
		return err
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
		return errs.MultiError(
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
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\n", ui.Bold("Policy Name"), ui.Bold("Type"), ui.Bold("Attached To"))
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

	policyName := args[0]
	if err := validate.PolicyName(policyName); err != nil {
		return errs.NewUsageExitError("Invalid policy name provided", ctx)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := getOrgWithPrompt(client, c, ctx.String("org"))
	if err != nil {
		return err
	}

	policies, err := client.Policies.List(c, org.ID, policyName)
	if err != nil {
		return errs.NewExitError("Unable to list policies.")
	}

	if len(policies) < 1 {
		return errs.NewExitError("Policy '" + policyName + "' not found.")
	}

	policy := policies[0]
	p := policy.Body.Policy

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\n", ui.Bold("Name:"), p.Name)
	fmt.Fprintf(w, "%s\t%s\n", ui.Bold("Description:"), p.Description)
	fmt.Fprintln(w, "")
	w.Flush()

	for _, stmt := range p.Statements {
		fmt.Fprintf(w, "%s\t%s\t%s\n", stmt.Effect.String(), stmt.Action.ShortString(), stmt.Resource)
	}
	w.Flush()

	return nil
}
