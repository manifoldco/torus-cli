package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
)

func init() {
	services := cli.Command{
		Name:     "services",
		Usage:    "Manage services within an organization",
		Category: "PROJECT STRUCTURE",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a service in an organization",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					orgFlag("Create the project in this org", false),
					projectFlag("project to create services for", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					createServiceCmd,
				),
			},
			{
				Name:  "list",
				Usage: "List services for an organization",
				Flags: []cli.Flag{
					orgFlag("org to show services for", true),
					projectFlag("project to show services for", false),
					cli.BoolFlag{
						Name:  "all",
						Usage: "Perform command on all projects",
					},
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, listServicesCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, services)
}

const serviceListFailed = "Could not list services."

func listServicesCmd(ctx *cli.Context) error {
	if !ctx.Bool("all") {
		if len(ctx.String("project")) < 1 {
			return errs.NewUsageExitError("Missing flags: --project", ctx)
		}
	} else {
		if len(ctx.String("project")) > 0 {
			return errs.NewUsageExitError("Cannot use --project flag with --all", ctx)
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	var org *envelope.Org
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(serviceListFailed, err)
	}
	if org == nil {
		return errs.NewExitError("Org not found")
	}

	// Identify which projects to list services for
	var projectID identity.ID
	var projects []envelope.Project
	if ctx.Bool("all") {
		// Pull all projects for the given orgID
		projects, err = listProjects(&c, client, org.ID, nil)
		if err != nil {
			return errs.NewErrorExitError(serviceListFailed, err)
		}

	} else {
		// Retrieve only a single project by name
		projectName := ctx.String("project")
		projects, err = listProjects(&c, client, org.ID, &projectName)
		if err != nil {
			return errs.NewErrorExitError(serviceListFailed, err)
		}
		if len(projects) == 1 {
			projectID = *projects[0].ID
		} else {
			return errs.NewExitError("Project not found")
		}
	}

	// Retrieve services for targeted org and project
	services, err := listServices(&c, client, org.ID, &projectID, nil)
	if err != nil {
		return errs.NewErrorExitError(serviceListFailed, err)
	}

	// Build map of services to project
	pMap := make(map[string]envelope.Project)
	for _, project := range projects {
		pMap[project.ID.String()] = project
	}
	sMap := make(map[string][]envelope.Service)
	for _, service := range services {
		ID := service.Body.ProjectID.String()
		sMap[ID] = append(sMap[ID], service)
	}

	// Build output of projects/services
	fmt.Println("")
	for projectID, project := range pMap {
		count := strconv.Itoa(len(sMap[projectID]))
		title := project.Body.Name + " (" + count + ")"
		fmt.Println(title)
		fmt.Println(strings.Repeat("-", utf8.RuneCountInString(title)))
		for _, service := range sMap[projectID] {
			fmt.Println(service.Body.Name)
		}
		fmt.Println("")
	}

	return nil
}

func listServices(ctx *context.Context, client *api.Client, orgID, projID *identity.ID, name *string) ([]envelope.Service, error) {
	c, client, err := NewAPIClient(ctx, client)
	if err != nil {
		return nil, cli.NewExitError(serviceListFailed, -1)
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

	return client.Services.List(c, orgIDs, projectIDs, names)
}

const serviceCreateFailed = "Could not create service."

func createServiceCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(serviceCreateFailed)
	}

	args := ctx.Args()
	serviceName := ""
	if len(args) > 0 {
		serviceName = args[0]
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
		return errs.NewExitError("Org not found")
	}
	if newOrg && oName == "" {
		fmt.Println("")
		return errs.NewExitError("Invalid org name")
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
		return errs.NewExitError("Project not found")
	}
	if newProject && pName == "" {
		fmt.Println("")
		return errs.NewExitError("Invalid project name")
	}

	label := "Service name"
	autoAccept := serviceName != ""
	serviceName, err = NamePrompt(&label, serviceName, autoAccept)
	if err != nil {
		return handleSelectError(err, serviceCreateFailed)
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

	// Create our new service
	fmt.Println("")
	err = client.Services.Create(c, orgID, project.ID, serviceName)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Service already exists")
		}
		return errs.NewErrorExitError(serviceCreateFailed, err)
	}

	fmt.Printf("Service %s created.\n", serviceName)
	return nil
}
