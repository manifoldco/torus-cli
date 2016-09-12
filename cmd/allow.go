package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/pathexp"
	"github.com/arigatomachine/cli/primitive"
)

func init() {
	allow := cli.Command{
		Name:      "allow",
		Usage:     "Grant a team permission to access specific resources",
		ArgsUsage: "<crudl> <path> <team>",
		Category:  "ACCESS CONTROL",
		Action:    chain(ensureDaemon, ensureSession, allowCmd),
	}

	Cmds = append(Cmds, allow)
}

func allowCmd(ctx *cli.Context) error {
	err := doCrudl(ctx, primitive.PolicyEffectAllow,
		primitive.PolicyActionList|primitive.PolicyActionRead)

	if err == nil {
		fmt.Println("\nNecessary permissions (read, list) have also been granted.")
	}

	return err
}

func doCrudl(ctx *cli.Context, effect primitive.PolicyEffect, extra primitive.PolicyAction) error {
	args := ctx.Args()
	if len(args) != 3 {
		msg := "permissions, path, and team are required.\n"
		if len(args) > 2 {
			msg = "Too many arguments provided.\n"
		}
		msg += usageString(ctx)
		return cli.NewExitError(msg, -1)
	}

	idx := strings.LastIndex(args[1], "/")
	if idx == -1 {
		msg := "resource path format is incorrect.\n"
		msg += usageString(ctx)
		return cli.NewExitError(msg, -1)
	}

	name := args[1][idx+1:]
	path := args[1][:idx]

	pe, err := pathexp.Parse(path)
	if err != nil {
		return cli.NewExitError("Error reading path expression: "+err.Error(), -1)
	}

	stmtAction, err := parseAction(args[0])
	if err != nil {
		return err
	}

	stmtAction |= extra

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := client.Orgs.GetByName(c, pe.Org())
	if err != nil {
		return cli.NewExitError("Unable to lookup org. Please try again.", -1)
	}
	if org == nil {
		return cli.NewExitError("Org not found.", -1)
	}

	teams, err := client.Teams.GetByName(c, org.ID, args[2])
	if err != nil {
		return cli.NewExitError("Unable to lookup team. Please try again.", -1)
	}
	if len(teams) < 1 {
		return cli.NewExitError("Team not found.", -1)
	}
	team := &teams[0]

	policy := primitive.Policy{
		PolicyType: "user",
		OrgID:      org.ID,
	}
	policy.Policy.Name = fmt.Sprintf("generated-%s-%d", effect.String(), time.Now().Unix())
	policy.Policy.Statements = []primitive.PolicyStatement{{
		Effect:   effect,
		Action:   stmtAction,
		Resource: pe.String() + "/" + name,
	}}

	res, err := client.Policies.Create(c, &policy)
	if err != nil {
		return cli.NewExitError("Error creating policy: "+err.Error(), -1)
	}

	err = client.Policies.Attach(c, org.ID, res.ID, team.ID)
	if err != nil {
		return cli.NewExitError("Could not attach policy: "+err.Error(), -1)
	}

	fmt.Printf("Policy generated and attached to the %s team.\n", team.Body.Name)

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	for _, s := range res.Body.Policy.Statements {
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "Effect:\t%s\n", s.Effect.String())
		fmt.Fprintf(w, "Action(s):\t%s\n", s.Action.String())
		fmt.Fprintf(w, "Resource:\t%s\n", s.Resource)
	}
	w.Flush()

	return nil
}

func parseAction(raw string) (primitive.PolicyAction, error) {
	var action primitive.PolicyAction
	for _, c := range raw {
		var add primitive.PolicyAction
		switch c {
		case 'c':
			add = primitive.PolicyActionCreate
		case 'r':
			add = primitive.PolicyActionRead
		case 'u':
			add = primitive.PolicyActionUpdate
		case 'd':
			add = primitive.PolicyActionDelete
		case 'l':
			add = primitive.PolicyActionList
		default:
			return action, cli.NewExitError("Unknown access character: "+string(c), -1)
		}

		if action&add > 0 {
			return action, cli.NewExitError("Duplication permission '"+string(c)+"' given.", -1)
		}

		action |= add
	}

	return action, nil
}
