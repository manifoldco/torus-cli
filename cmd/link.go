package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/dirprefs"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"
)

func init() {
	link := cli.Command{
		Name:     "link",
		Usage:    "Link your current directory to Torus",
		Category: "PROJECT STRUCTURE",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Overwrite existing organization and project links.",
			},
		},
		Action: chain(ensureDaemon, ensureSession, linkCmd),
	}

	Cmds = append(Cmds, link)
}

func linkCmd(ctx *cli.Context) error {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return err
	}

	dPrefs, err := dirprefs.Load(false)
	if err != nil {
		return err
	}

	if dPrefs != nil && dPrefs.Path != "" && !ctx.Bool("force") {
		msg := fmt.Sprintf(
			"This directory is already linked. Use '%s status' to view.",
			ctx.App.Name,
		)
		return errs.NewExitError(msg)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, oName, newOrg, err := selectOrg(c, client, ctx.String("org"), true)
	if err != nil {
		return err
	}

	var orgID *identity.ID
	if !newOrg {
		orgID = org.ID
	}

	project, pName, newProject, err := selectProject(c, client, org, ctx.String("project"), true)
	if err != nil {
		return err
	}

	// Create the org now if needed
	if newOrg {
		org, err = createOrgByName(c, ctx, client, oName)
		if err != nil {
			fmt.Println("")
			return err
		}
		orgID = org.ID
	}

	// Create the project now if needed
	if newProject {
		project, err = createProjectByName(c, client, orgID, pName)
		if err != nil {
			fmt.Println("")
			return err
		}
	}

	// write out the link
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	oName = org.Body.Name
	pName = project.Body.Name

	dPrefs.Organization = oName
	dPrefs.Project = pName
	dPrefs.Path = filepath.Join(cwd, ".torus.json")

	err = dPrefs.Save()
	if err != nil {
		return err
	}

	// Display the output
	fmt.Println("\nThis directory and its subdirectories have been linked to:")
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Org"), oName)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Project"), pName)
	w.Flush()
	fmt.Printf("\nUse '%s status' to view your full working context.\n", ctx.App.Name)

	if !preferences.Core.Context {
		ui.Warn(fmt.Sprintf("Context is disaled. Use '%s prefs' to enable it.", ctx.App.Name))
	}

	hints.Display(hints.Context, hints.Set, hints.Run, hints.View)
	return nil
}
