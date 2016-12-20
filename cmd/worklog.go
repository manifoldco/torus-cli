package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/promptui"
)

var catOrder = []apitypes.WorklogType{
	apitypes.MissingKeypairsWorklogType,
	apitypes.InviteApproveWorklogType,
	apitypes.UserKeyringMembersWorklogType,
	apitypes.MachineKeyringMembersWorklogType,
	apitypes.SecretRotateWorklogType,
}

var (
	yellow = promptui.Styler(promptui.FGYellow)

	faint     = promptui.Styler(promptui.FGFaint)
	underline = promptui.Styler(promptui.FGUnderline)
	italic    = promptui.Styler(promptui.FGItalic)
)

func init() {
	worklog := cli.Command{
		Name:     "worklog",
		Usage:    "View and perform maintenance tasks",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List worklog maintenance tasks",
				Flags: []cli.Flag{stdOrgFlag},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, worklogList,
				),
			},
			{
				Name:      "view",
				Usage:     "Show the details of a worklog item",
				ArgsUsage: "<identity>",
				Flags:     []cli.Flag{stdOrgFlag},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, worklogView,
				),
			},
			{
				Name:      "resolve",
				Usage:     "Act on and resolve the given worklog items",
				ArgsUsage: "[identity...]",
				Flags:     []cli.Flag{stdOrgFlag},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, worklogResolve,
				),
			},
		},
	}
	Cmds = append(Cmds, worklog)
}

func groupMsgFor(typ apitypes.WorklogType) string {
	switch typ {
	case apitypes.MissingKeypairsWorklogType:
		return "Orgs with missing keypairs:"
	case apitypes.InviteApproveWorklogType:
		return "Invites ready for approval to the %s org:"
	case apitypes.UserKeyringMembersWorklogType:
		return "Users missing granted access to secrets in the %s org:"
	case apitypes.MachineKeyringMembersWorklogType:
		return "Machines missing granted access to secrets in the %s org:"
	case apitypes.SecretRotateWorklogType:
		return "Secrets that should be rotated in the %s org:"
	default:
		return ""
	}
}

func subjectFor(item *apitypes.WorklogItem) string {
	switch d := item.Details.(type) {
	case *apitypes.MissingKeypairsWorklogDetails:
		return underline(d.Org)
	case *apitypes.InviteApproveWorklogDetails:
		return fmt.Sprintf("%s <%s>", underline(d.Username), italic(d.Email))
	case *apitypes.KeyringMembersWorklogDetails:
		return underline(d.Name)
	case *apitypes.SecretRotateWorklogDetails:
		return item.Subject()
	default:
		return item.Subject()
	}
}

func detailsFor(org *envelope.Org, item *apitypes.WorklogItem) string {
	switch d := item.Details.(type) {
	case *apitypes.MissingKeypairsWorklogDetails:
		return underline(d.Org) + "\n"
	case *apitypes.InviteApproveWorklogDetails:
		msg := fmt.Sprintf("  The invite for %s to the %s org is ready for approval.\n",
			d.Name, underline(org.Body.Name))
		msg += "  They will be invited to the following teams:\n"
		for _, t := range d.Teams {
			msg += fmt.Sprintf("    %s\n", t)
		}
		return msg
	case *apitypes.KeyringMembersWorklogDetails:
		msg := fmt.Sprintf("  %s is missing granted access to secrets in the %s org.\n",
			underline(d.Name), underline(org.Body.Name))
		msg += "  Secrets in the following paths are affected:\n"
		for _, p := range d.Keyrings {
			msg += fmt.Sprintf("    %s\n", p.String())
		}
		return msg
	case *apitypes.SecretRotateWorklogDetails:
		msg := "  The value for this secret should be rotated for the following reasons:\n"
		for _, r := range d.Reasons {
			var rm string
			switch r.Type {
			case primitive.OrgRemovalRevocationType:
				rm = "was removed from the org."
			case primitive.KeyRevocationRevocationType:
				rm = "changed their encryption key."
			default:
				rm = "lost access."
			}

			msg += fmt.Sprintf("    %s %s\n", underline(r.Username), rm)
		}
		return msg
	default:
		return item.Subject() + "\n"
	}
}

func worklogList(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := getOrg(c, client, ctx.String("org"))
	if err != nil {
		return err
	}

	items, err := client.Worklog.List(c, org.ID)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("Worklog complete! No items left to resolve. 👍")
		return nil
	}

	itemsByCat := make(map[apitypes.WorklogType][]apitypes.WorklogItem)
	for _, item := range items {
		itemsByCat[item.Type()] = append(itemsByCat[item.Type()], item)
	}

	newlineNeeded := false

	for _, cat := range catOrder {
		items = itemsByCat[cat]
		if len(items) == 0 {
			continue
		}

		if newlineNeeded {
			fmt.Println()
		}
		newlineNeeded = true

		groupMsg := fmt.Sprintf(groupMsgFor(cat), underline(org.Body.Name))
		fmt.Printf("%s %s\n", yellow(cat.String()), groupMsg)
		for _, item := range items {
			fmt.Printf("  %s %s\n", faint(item.ID.String()), subjectFor(&item))
		}
	}

	return nil
}

func worklogView(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 1 {
		msg := "Identity is required."
		if len(args) > 2 {
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

	org, err := getOrg(c, client, ctx.String("org"))
	if err != nil {
		return err
	}

	ident, err := apitypes.DecodeWorklogIDFromString(args[0])
	if err != nil {
		return errs.NewExitError("Malformed id for worklog item.")
	}

	item, err := client.Worklog.Get(c, org.ID, &ident)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n", yellow(item.ID.String()), subjectFor(item))
	fmt.Printf(detailsFor(org, item))
	return nil
}

func worklogResolve(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := getOrg(c, client, ctx.String("org"))
	if err != nil {
		return err
	}

	var idents []apitypes.WorklogID
	for _, raw := range ctx.Args() {
		ident, err := apitypes.DecodeWorklogIDFromString(raw)
		if err != nil {
			return errs.NewExitError("Malformed id for worklog item.")
		}

		idents = append(idents, ident)
	}

	items, err := client.Worklog.List(c, org.ID)
	if err != nil {
		return err
	}

	var toResolve []apitypes.WorklogItem
	if len(idents) == 0 {
		toResolve = items
	} else {
	IdentLoop:
		for _, ident := range idents {
			for _, item := range items {
				if *item.ID == ident {
					toResolve = append(toResolve, item)
					continue IdentLoop
				}
			}

			return errs.NewExitError("Could not find worklog item with identity " + ident.String())
		}
	}

	itemsByCat := make(map[apitypes.WorklogType][]apitypes.WorklogItem)
	for _, item := range toResolve {
		itemsByCat[item.Type()] = append(itemsByCat[item.Type()], item)
	}

	grouped := len(idents) == 0 // no manual ids; doing all of them

	newlineNeeded := false

	for _, cat := range catOrder {
		items = itemsByCat[cat]
		if len(items) == 0 {
			continue
		}

		if grouped {
			if newlineNeeded {
				fmt.Println()
			}
			newlineNeeded = true

			groupMsg := fmt.Sprintf(groupMsgFor(cat), underline(org.Body.Name))
			fmt.Printf("%s %s\n", yellow(cat.String()), groupMsg)
		}

		for _, item := range items {
			// An explicit invite id won't trigger a prompt
			if item.Type() == apitypes.InviteApproveWorklogType && grouped {
				msg := fmt.Sprintf("%s%s Approve invite for %s", promptui.ResetCode,
					faint(item.ID.String()), subjectFor(&item))
				err = AskPerform(msg, "  ")
				switch err {
				case nil:
				case promptui.ErrAbort:
					continue
				default:
					return err
				}
			} else if item.Type() == apitypes.SecretRotateWorklogType {
				displayResult(&item, nil, grouped)
				continue
			}

			res, err := client.Worklog.Resolve(c, org.ID, item.ID)
			if err == nil && res.State != apitypes.SuccessWorklogResult {
				err = errors.New(res.Message)
			}
			displayResult(&item, err, grouped)
		}
	}

	return nil
}

func displayResult(item *apitypes.WorklogItem, err error, grouped bool) {
	icon := promptui.IconGood

	if item.Type() == apitypes.SecretRotateWorklogType {
		icon = promptui.IconWarn
	}

	indent := ""
	idFmt := yellow

	if grouped {
		indent = "  "
		idFmt = faint
	}

	var message string
	if err != nil {
		icon = promptui.IconBad

		var typ string
		switch item.Type() {
		case apitypes.MissingKeypairsWorklogType:
			typ = "generating keypairs"
		case apitypes.InviteApproveWorklogType:
			typ = "approving invite"
		case apitypes.UserKeyringMembersWorklogType:
			fallthrough
		case apitypes.MachineKeyringMembersWorklogType:
			typ = "reconciling secret access"
		case apitypes.SecretRotateWorklogType:
			typ = "rotating secret" // this one will never happen; its manual.
		}

		message = fmt.Sprintf("Error %s: %s", typ, err)
	} else {
		switch item.Type() {
		case apitypes.MissingKeypairsWorklogType:
			message = "Keypairs generated for %s"
		case apitypes.InviteApproveWorklogType:
			message = "Invite approved for %s"
		case apitypes.UserKeyringMembersWorklogType:
			message = "Secret access for user %s has been reconciled."
		case apitypes.MachineKeyringMembersWorklogType:
			message = "Secret access for machine %s has been reconciled."
		case apitypes.SecretRotateWorklogType:
			message = "Please set a new value for\n" + indent + "    %s"
		}

		message = fmt.Sprintf(message, subjectFor(item))
	}

	fmt.Printf("%s%s %s %s\n", indent, icon, idFmt(item.ID.String()), message)
}
