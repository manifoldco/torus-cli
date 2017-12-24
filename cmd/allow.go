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
	"github.com/manifoldco/torus-cli/validate"
)

func init() {
	allow := cli.Command{
		Name:      "allow",
		Usage:     "Increase access given to a team or role by creating and attaching a new policy",
		ArgsUsage: "<crudl> <path> <team|machine-role>",
		Category:  "ACCESS CONTROL",
		Flags: []cli.Flag{
			nameFlag("The name to give the generated policy (e.g. allow-prod-env)"),
			descriptionFlag("A sentence or two explaining the purpose of the policy"),
		},
		Action: chain(ensureDaemon, ensureSession, allowCmd),
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
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	args := ctx.Args()
	if len(args) != 3 {
		msg := "permissions, path, and team are required."
		if len(args) > 2 {
			msg = "Too many arguments provided."
		}
		return errs.NewUsageExitError(msg, ctx)
	}

	// Separate the pathexp from the secret name
	pe, secretName, err := parseRawPath(args[1])
	if err != nil {
		return err
	}

	stmtAction, err := parseAction(args[0])
	if err != nil {
		return err
	}

	stmtAction |= extra

	name := ctx.String("name")
	description := ctx.String("description")

	if name == "" {
		name = fmt.Sprintf("generated-%s-%d", effect.String(), time.Now().Unix())
	}
	if description == "" {
		session, err := client.Session.Who(c)
		if err != nil {
			return errs.NewErrorExitError("Error fetching identity", err)
		}

		description = fmt.Sprintf("Generated on %s by %s",
			time.Now().Format(time.RFC822Z), session.Username())
	}

	if err := validate.PolicyName(name); err != nil {
		return errs.NewErrorExitError("Invalid name provided.", err)
	}

	if err := validate.Description(description, "policy"); err != nil {
		return errs.NewErrorExitError("Invalid description provided.", err)
	}

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
	policy.Policy.Name = name
	policy.Policy.Description = description
	policy.Policy.Statements = []primitive.PolicyStatement{{
		Effect:   effect,
		Action:   stmtAction,
		Resource: pe.String() + "/" + *secretName,
	}}

	res, err := client.Policies.Create(c, &policy)
	if err != nil {
		return errs.NewErrorExitError("Failed to create policy", err)
	}

	err = client.Policies.Attach(c, org.ID, res.ID, team.ID)
	if err != nil {
		return errs.NewErrorExitError("Could not attach policy.", err)
	}

	fmt.Printf("Policy %s generated and attached to the %s team.\n", name, team.Body.Name)

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", name)
	fmt.Fprintf(w, "Description:\t%s\n", description)

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
