package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/promptui"
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

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, "IDENTITY\tTYPE\tSUBJECT")
	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\n", item.ID, item.Type(), item.Subject)
	}

	w.Flush()
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

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Identity:\t%s\n", item.ID.String())
	fmt.Fprintf(w, "Type:\t%s\n", item.Type())
	fmt.Fprintf(w, "Subject:\t%s\n", item.Subject)
	w.Flush()
	fmt.Println("Summary:")
	fmt.Println(item.Summary)

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

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	for _, item := range toResolve {
		if item.Type() == apitypes.InviteApproveWorklogType {
			w.Flush()

			err = AskPerform("Approve invite for " + item.Subject)
			switch err {
			case nil:
			case promptui.ErrAbort:
				continue
			default:
				return err
			}
		}

		res, err := client.Worklog.Resolve(c, org.ID, item.ID)
		if err != nil {
			return errs.NewErrorExitError("Error resolving worklog item.", err)
		}

		var icon string
		switch res.State {
		case apitypes.SuccessWorklogResult:
			icon = "✔"
		case apitypes.FailureWorklogResult:
			icon = "✗"
		case apitypes.ManualWorklogResult:
			icon = "⚠"
		default:
			icon = "?"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", icon, res.ID, res.Message)
	}
	w.Flush()

	return nil
}
