package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
)

func init() {
	allow := cli.Command{
		Name:      "allow",
		Usage:     "Increase access given to a team or role by creating and attaching a new policy",
		ArgsUsage: "<crudl> <path> <team|machine-role>",
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
		msg := "permissions, path, and team are required."
		if len(args) > 2 {
			msg = "Too many arguments provided."
		}
		return errs.NewUsageExitError(msg, ctx)
	}

	// Separate the pathexp from the secret name
	pe, name, err := parseRawPath(args[1])
	if err != nil {
		return err
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

	org, err := client.Orgs.GetByName(c, pe.Org.String())
	if err != nil {
		return errs.NewErrorExitError("Unable to lookup org.", err)
	}
	if org == nil {
		return errs.NewExitError("Org not found")
	}

	teams, err := client.Teams.GetByName(c, org.ID, args[2])
	if err != nil {
		return errs.NewErrorExitError("Unable to lookup team.", err)
	}
	if len(teams) < 1 {
		return errs.NewExitError("Team not found.")
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
		Resource: pe.String() + "/" + *name,
	}}

	res, err := client.Policies.Create(c, &policy)
	if err != nil {
		return errs.NewErrorExitError("Failed to create policy", err)
	}

	err = client.Policies.Attach(c, org.ID, res.ID, team.ID)
	if err != nil {
		return errs.NewErrorExitError("Could not attach policy.", err)
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

// parseRawPath parses and validates a raw path, possibly including a secret.
// The secret portion is also validated.
func parseRawPath(rawPath string) (*pathexp.PathExp, *string, error) {
	idx := strings.LastIndex(rawPath, "/")
	if idx == -1 {
		return nil, nil, errs.NewExitError("resource path format is incorrect.")
	}
	secret := rawPath[idx+1:]
	path := rawPath[:idx]

	if secret == "**" {
		path = rawPath
		secret = "*"
	}

	// Ensure that the secret name is valid
	if !pathexp.ValidSecret(secret) {
		return nil, nil, errs.NewExitError("Invalid secret name " + secret)
	}

	pe, err := pathexp.Parse(path)
	if err != nil {
		return nil, nil, errs.NewErrorExitError("Invalid path expression", err)
	}
	return pe, &secret, nil
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
			return action, errs.NewExitError("Unknown access character: " + string(c))
		}

		if action&add > 0 {
			return action, errs.NewExitError("Duplication permission '" + string(c) + "' given.")
		}

		action |= add
	}

	return action, nil
}
