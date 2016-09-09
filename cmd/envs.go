package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/promptui"
)

func init() {
	envs := cli.Command{
		Name:     "envs",
		Usage:    "View and manipulate environments within an organization",
		Category: "ORGANIZATIONS",
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
					orgFlag("org to show environments for", true),
					projectFlag("project to shows environments for", false),
					cli.BoolFlag{
						Name:  "all",
						Usage: "Perform command on all projects",
					},
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, listEnvs,
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
		return cli.NewExitError(envCreateFailed, -1)
	}

	args := ctx.Args()
	environmentName := ""
	if len(args) > 0 {
		environmentName = args[0]
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
		fmt.Println("")
		return cli.NewExitError("Project not found", -1)
	}
	if newProject && pName == "" {
		fmt.Println("")
		return cli.NewExitError("Invalid project name", -1)
	}

	label := "Environment name"
	if environmentName == "" {
		environmentName, err = NamePrompt(&label, "")
		if err != nil {
			return handleSelectError(err, envCreateFailed)
		}
	} else {
		fmt.Println(promptui.SuccessfulValue(label, environmentName))
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

	// Create our new environment
	fmt.Println("")
	err = client.Environments.Create(c, orgID, project.ID, environmentName)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return cli.NewExitError("Environment already exists", -1)
		}
		return cli.NewExitError(envCreateFailed, -1)
	}

	fmt.Println("Environment " + environmentName + " created.")
	return nil
}

const envListFailed = "Could not list envs, please try again."

func listEnvs(ctx *cli.Context) error {
	if !ctx.Bool("all") {
		if len(ctx.String("project")) < 1 {
			text := "Missing flags: --project\n"
			text += usageString(ctx)
			return cli.NewExitError(text, -1)
		}
	}
	// TODO: Error when profile flag is used with --all

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	var org *api.OrgResult
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return cli.NewExitError(envListFailed, -1)
	}
	if org == nil {
		return cli.NewExitError("Org not found.", -1)
	}

	// Identify which projects to list envs for
	var projectID identity.ID
	var projects []api.ProjectResult
	if ctx.Bool("all") {
		// Pull all projects for the given orgID
		projects, err = client.Projects.List(c, org.ID, nil)
		if err != nil {
			return cli.NewExitError(envListFailed, -1)
		}

	} else {
		// Retrieve only a single project by name
		projectName := ctx.String("project")
		projects, err = client.Projects.List(c, org.ID, &projectName)
		if err != nil {
			return cli.NewExitError(envListFailed, -1)
		}
		if len(projects) == 1 {
			projectID = *projects[0].ID
		} else {
			return cli.NewExitError("Project not found.", -1)
		}
	}

	// Retrieve envs for targeted org and project
	var envs []api.EnvironmentResult
	envs, err = client.Environments.List(c, org.ID, &projectID, nil)
	if err != nil {
		fmt.Printf("%v", err.Error())
		return cli.NewExitError(envListFailed, -1)
	}

	// Build map of envs to project
	pMap := make(map[string]api.ProjectResult)
	for _, project := range projects {
		pMap[project.ID.String()] = project
	}
	eMap := make(map[string][]api.EnvironmentResult)
	for _, env := range envs {
		ID := env.Body.ProjectID.String()
		eMap[ID] = append(eMap[ID], env)
	}

	// Build output of projects/envs
	fmt.Println("")
	for projectID, project := range pMap {
		count := strconv.Itoa(len(eMap[projectID]))
		title := project.Body.Name + " (" + count + ")"
		fmt.Println(title)
		fmt.Println(strings.Repeat("-", utf8.RuneCountInString(title)))
		for _, env := range eMap[projectID] {
			fmt.Println(env.Body.Name)
		}
		fmt.Println("")
	}

	return nil
}
