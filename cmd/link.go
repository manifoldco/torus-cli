package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/dirprefs"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/prefs"
)

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
		return cli.NewExitError(msg, -1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Ask the user which org they want to use
	org, oName, newOrg, err := SelectCreateOrg(client, c, ctx.String("org"))
	if err != nil {
		return handleSelectError(err, "Org selection failed")
	}
	if org == nil && !newOrg {
		fmt.Println("")
		return cli.NewExitError("Org not found", -1)
	}
	if newOrg && oName == "" {
		fmt.Println("")
		return cli.NewExitError("Invalid org name", -1)
	}

	var orgID *identity.ID
	if org != nil {
		orgID = org.ID
	}

	// Ask the user which project they want to use
	project, pName, newProject, err := SelectCreateProject(client, c, orgID, ctx.String("project"))
	if err != nil {
		return handleSelectError(err, "Project selection failed")
	}
	if project == nil && !newProject {
		return cli.NewExitError("Project not found", -1)
	}
	if newProject && pName == "" {
		return cli.NewExitError("Invalid project name", -1)
	}

	// Create the org now if needed
	if org == nil && newOrg {
		org, err = createOrgByName(ctx, c, client, oName)
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
			return cli.NewExitError("Service creation failed", -1)
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
