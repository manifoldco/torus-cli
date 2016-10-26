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
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/prefs"
)

func init() {
	link := cli.Command{
		Name:     "link",
		Usage:    "Link your current directory to Torus",
		Category: "CONTEXT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Overwrite existing organization and project links.",
			},
			cli.BoolFlag{
				Name:   "bare",
				Usage:  "Skip creation of default service.",
				Hidden: true,
			},
		},
		Action: chain(ensureDaemon, ensureSession, linkCmd),
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
		return errs.NewExitError(msg)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Ask the user which org they want to use
	org, oName, newOrg, err := SelectCreateOrg(c, client, ctx.String("org"))
	if err != nil {
		return handleSelectError(err, "Org selection failed.")
	}
	if org == nil && !newOrg {
		fmt.Println("")
		return errs.NewExitError("Org not found.")
	}
	if newOrg && oName == "" {
		fmt.Println("")
		return errs.NewExitError("Invalid org name.")
	}

	var orgID *identity.ID
	if org != nil {
		orgID = org.ID
	}

	// Ask the user which project they want to use
	project, pName, newProject, err := SelectCreateProject(c, client, orgID, ctx.String("project"))
	if err != nil {
		return handleSelectError(err, "Project selection failed.")
	}
	if project == nil && !newProject {
		return errs.NewExitError("Project not found.")
	}
	if newProject && pName == "" {
		return errs.NewExitError("Invalid project name.")
	}

	// Create the org now if needed
	if org == nil && newOrg {
		org, err = createOrgByName(c, ctx, client, oName)
		if err != nil {
			fmt.Println("")
			return err
		}
		orgID = org.ID
	}

	// Create the project now if needed
	if project == nil && newProject {
		project, err = createProjectByName(c, client, orgID, pName)
		if err != nil {
			fmt.Println("")
			return err
		}
	}

	// Do not create default service if --bare is used or if existing project
	// is selected from the prompt
	createDefaultService := true
	if ctx.Bool("bare") || !newProject {
		createDefaultService = false
	}

	// Create the default service if necessary
	if createDefaultService {
		err = client.Services.Create(c, org.ID, project.ID, "default")
		if err != nil {
			fmt.Println("")
			return errs.NewExitError("Service creation failed.")
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
