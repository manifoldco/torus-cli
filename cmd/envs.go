package cmd

import (
	"context"
	"fmt"
	"strconv"
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

	var org *envelope.Org
	var orgs []envelope.Org

	// Retrieve the org name supplied via the --org flag.
	// This flag is optional. If none was supplied, then
	// orgFlagArgument will be set to "". In this case,
	// prompt the user to select an org.
	orgName := ctx.String("org")

	if orgName == "" {
		// Retrieve list of available orgs
		orgs, err = client.Orgs.List(c)
		if err != nil {
			return errs.NewErrorExitError("Failed to retrieve orgs list.", err)
		}

		// Prompt user to select from list of existing orgs
		idx, _, err := SelectExistingOrgPrompt(orgs)
		if err != nil {
			return errs.NewErrorExitError("Failed to select org.", err)
		}

		org = &orgs[idx]
		orgName = org.Body.Name

	} else {
		// If org flag was used, identify the org supplied.
		org, err = client.Orgs.GetByName(c, orgName)
		if err != nil {
			return errs.NewErrorExitError("Failed to retrieve org " + orgName, err)
		}
		if org == nil {
			return errs.NewExitError("org " + orgName + " not found.")
		}
	}

	// Retrieve project by name
	// If no project was supplied (via flags) prompt the
	// user to select from a list of existing projects.
	var project *envelope.Project
	var projects []envelope.Project

	projectName := ctx.String("project")

	if projectName == "" {
		// Retrieve list of available projects
		projects, err = client.Projects.List(c, org.ID)
		if err != nil {
			return errs.NewErrorExitError("Failed to retrieve projects list.", err)
		}

		// Prompt user to select from list of existing orgs
		idx, _, err := SelectExistingProjectPrompt(projects)
		if err != nil {
			return errs.NewErrorExitError("Failed to select project.", err)
		}

		project = &projects[idx]
		projectName = project.Body.Name

	} else {
		// Get Project for project name, confirm project exists
		project, err = getProject(c, client, org.ID, projectName)
		if err != nil {
			return errs.NewErrorExitError("Project " + projectName + " not found.", err)
		}
	}

	// Retrieve envs for targeted org and project
	envs, err := listEnvs(&c, client, org.ID, project.ID, nil)
	if err != nil {
		return errs.NewErrorExitError(envListFailed, err)
	}

	// Build output of projects/envs
	fmt.Println("")
	fmt.Printf("%s\n", ui.Bold("Environments"))
	for _, env := range envs {
		fmt.Printf("%s\n", env.Body.Name)
	}
	fmt.Println("")

	count := strconv.Itoa(len(envs))
	title := "Project /" + org.Body.Name + "/" + project.Body.Name + " has (" + count + ") environments."
	fmt.Println(title)
	fmt.Println("")

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
