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

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/identity"
)

func init() {
	policies := cli.Command{
		Name:     "policies",
		Usage:    "View and manipulate access control list policies",
		Category: "ACCESS CONTROL",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List services for an organization",
				Flags: []cli.Flag{
					OrgFlag("org to show services for", true),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, listPolicies,
				),
			},
		},
	}
	Cmds = append(Cmds, policies)
}

const policyListFailed = "Could not list policies, please try again."

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
		return cli.NewExitError(policyListFailed, -1)
	}
	if org == nil {
		return cli.NewExitError("Org not found.", -1)
	}

	var getAttachments, display sync.WaitGroup
	getAttachments.Add(3)
	display.Add(1)

	var policies []api.PoliciesResult
	var pErr error
	go func() {
		policies, pErr = client.Policies.List(c, org.ID, nil)
		getAttachments.Done()
	}()

	var attachments []api.PolicyAttachmentResult
	var aErr error
	go func() {
		attachments, aErr = client.Policies.AttachmentsList(c, org.ID)
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
			cli.NewExitError(policyListFailed, -1),
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
