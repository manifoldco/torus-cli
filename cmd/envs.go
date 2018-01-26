package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/ui"
)

func init() {
	envs := cli.Command{
		Name:     "envs",
		Usage:    "Manage environments within an organization",
		Category: "PROJECT STRUCTURE",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create an environment for a service inside an organization or project",
				Flags: []cli.Flag{
					orgFlag("org to create environment for", false),
					projectFlag("project to create environment for", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, createEnv,
				),
			},
			{
				Name:  "list",
				Usage: "List environments for an organization",
				Flags: []cli.Flag{
					orgFlag("org to show environments for", false),
					projectFlag("project to show environments for", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, listEnvsCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, envs)
}

const envCreateFailed = "Could not create environment, please try again."

func createEnv(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(envCreateFailed)
	}

	args := ctx.Args()
	environmentName := ""
	if len(args) > 0 {
		environmentName = args[0]
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
		fmt.Println("")
		return errs.NewExitError("Project not found.")
	}
	if newProject && pName == "" {
		fmt.Println("")
		return errs.NewExitError("Invalid project name.")
	}

	label := "Environment name"
	autoAccept := environmentName != ""
	environmentName, err = NamePrompt(&label, environmentName, autoAccept)
	if err != nil {
		return handleSelectError(err, envCreateFailed)
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

	// Create our new environment
	fmt.Println("")
	err = client.Environments.Create(c, orgID, project.ID, environmentName)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Environment already exists.")
		}
		return errs.NewExitError(envCreateFailed)
	}

	fmt.Println("Environment " + environmentName + " created.")
	return nil
}

const envListFailed = "Could not list envs, please try again."

func listEnvsCmd(ctx *cli.Context) error {

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := getOrgWithPrompt(c, client, ctx.String("org"))
	if err != nil {
		return err
	}

	project, err := getProjectWithPrompt(c, client, org, ctx.String("project"))
	if err != nil {
		return err
	}

	// Retrieve envs for targeted org and project
	envs, err := listEnvs(&c, client, org.ID, project.ID, nil)
	if err != nil {
		return errs.NewErrorExitError(envListFailed, err)
	}

	// Build output of projects/envs
	fmt.Println("")
	fmt.Printf("%s\n", ui.BoldString("Environments"))
	for _, env := range envs {
		fmt.Printf("%s\n", env.Body.Name)
	}

	fmt.Printf("\nProject /%s/%s has (%d) environment%s\n", org.Body.Name, project.Body.Name, len(envs), plural(len(envs)))

	return nil
}

func listEnvs(ctx *context.Context, client *api.Client, orgID, projID *identity.ID, name *string) ([]envelope.Environment, error) {
	c, client, err := NewAPIClient(ctx, client)
	if err != nil {
		return nil, cli.NewExitError(envListFailed, -1)
	}

	var orgIDs []identity.ID
	if orgID != nil {
		orgIDs = []identity.ID{*orgID}
	}

	var projectIDs []identity.ID
	if projID != nil {
		projectIDs = []identity.ID{*projID}
	}

	var names []string
	if name != nil {
		names = []string{*name}
	}

	return client.Environments.List(c, orgIDs, projectIDs, names)
}
