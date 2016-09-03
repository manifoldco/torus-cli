package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/dirprefs"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/prefs"
	"github.com/arigatomachine/cli/promptui"
)

var slug = regexp.MustCompile("^[a-z0-9][a-z0-9\\-\\_]{0,63}$")

func init() {
	link := cli.Command{
		Name:     "link",
		Usage:    "Link your current directory to Arigato",
		Category: "CONTEXT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Overwrite existing organization and project links.",
			},
		},
		Action: Chain(EnsureDaemon, EnsureSession, linkCmd),
	}

	Cmds = append(Cmds, link)
}

func linkCmd(ctx *cli.Context) error {
	preferences, err := prefs.NewPreferences(true)
	if err != nil {
		return err
	}

	dPrefs, err := dirprefs.Load(false)
	if dPrefs != nil && dPrefs.Path != "" && !ctx.Bool("force") {
		msg := fmt.Sprintf(
			"This directory is already linked. Use '%s status' to view.",
			ctx.App.Name,
		)
		return cli.NewExitError(msg, -1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	c := context.Background()

	// Get the list of orgs the user has access to
	orgs, err := client.Orgs.List(c)
	if err != nil {
		return cli.NewExitError("Error fetching orgs list", -1)
	}

	names := make([]string, len(orgs))
	for i, o := range orgs {
		names[i] = o.Body.Name
	}

	// Get the user's org selection
	prompt := promptui.SelectWithAdd{
		Label:    "Link to organization",
		Items:    names,
		AddLabel: "Create a new organization",
		Validate: validateSlug("org"),
	}

	oIdx, oName, err := prompt.Run()
	if err != nil {
		return err
	}

	names = []string{}
	var orgID *identity.ID

	// Load existing projects for the selected org
	if oIdx != promptui.SelectedAdd {
		orgID = orgs[oIdx].ID
		projects, err := client.Projects.List(c, orgID, nil)
		if err != nil {
			return cli.NewExitError("Error fetching projects list", -1)
		}

		names = make([]string, len(projects))
		for i, p := range projects {
			names[i] = p.Body.Name
		}
	}

	// Get the user's project selection
	prompt = promptui.SelectWithAdd{
		Label:    "Link to project",
		Items:    names,
		AddLabel: "Create a new project",
		Validate: validateSlug("project"),
	}

	pIdx, pName, err := prompt.Run()
	if err != nil {
		return err
	}

	if oIdx == promptui.SelectedAdd || pIdx == promptui.SelectedAdd {
		fmt.Println("")
	}

	// Create org if required
	if oIdx == promptui.SelectedAdd {
		org, err := client.Orgs.Create(c, oName)
		if err != nil {
			return cli.NewExitError("Could not create org: "+err.Error(), -1)
		}

		var progress api.ProgressFunc = func(evt *api.Event, err error) {
			if evt != nil {
				fmt.Println(evt.Message)
			}
		}

		orgID = org.ID
		err = client.Keypairs.Generate(c, orgID, &progress)
		if err != nil {
			msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
			return cli.NewExitError(msg, -1)
		}

		fmt.Printf("Org %s created.\n", oName)
	}

	// Create project if required
	if pIdx == promptui.SelectedAdd {
		_, err := client.Projects.Create(c, orgID, pName)
		if err != nil {
			return cli.NewExitError("Could not create project: "+err.Error(), -1)
		}

		fmt.Printf("Project %s created.\n", pName)
	}

	// write out the link
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dPrefs.Organization = oName
	dPrefs.Project = pName
	dPrefs.Path = filepath.Join(cwd, ".arigato.json")

	err = dPrefs.Save()
	if err != nil {
		return err
	}

	// Display the output
	fmt.Println("\nThis directory its subdirectories have been linked to:")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Org:\t%s\n", oName)
	fmt.Fprintf(w, "Project:\t%s\n", pName)
	w.Flush()
	fmt.Printf("\nUse '%s status' to view your full working context.\n", ctx.App.Name)

	if !preferences.Core.Context {
		fmt.Printf("Warning: context is disabled. Use '%s prefs' to enable it.\n", ctx.App.Name)
	}

	return nil
}

func validateSlug(slugType string) promptui.ValidateFunc {
	msg := slugType + " names can only use a-z, 0-9, hyphens and underscores"
	err := promptui.NewValidationError(msg)
	return func(input string) error {
		if !slug.Match([]byte(input)) {
			return err
		}
		return nil
	}
}
