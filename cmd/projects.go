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
	projects := cli.Command{
		Name:     "projects",
		Usage:    "Manage projects within an organization",
		Category: "PROJECT STRUCTURE",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a project in an organization",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					orgFlag("Create the project in this org", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					createProjectCmd,
				),
			},
			{
				Name:  "list",
				Usage: "List services for an organization",
				Flags: []cli.Flag{
					orgFlag("List projects in an organization", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, listProjectsCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, projects)
}

const projectListFailed = "Could not list projects, please try again."

func listProjectsCmd(ctx *cli.Context) error {

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := getOrgWithPrompt(client, c, ctx.String("org"))
	if err != nil {
		return err
	}

	projects, err := client.Projects.List(c, org.ID)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve projects list.", err)
	}

	fmt.Println("")
	fmt.Printf("%s\n", ui.Bold("Projects"))
	for _, project := range projects {
		fmt.Printf("%s\n", project.Body.Name)
	}

	fmt.Printf("\nOrg %s has (%d) project%s\n", org.Body.Name, len(projects), plural(len(projects)))

	return nil
}

func listProjects(ctx *context.Context, client *api.Client, orgID *identity.ID, name *string) ([]envelope.Project, error) {
	c, client, err := NewAPIClient(ctx, client)
	if err != nil {
		return nil, cli.NewExitError(projectListFailed, -1)
	}

	var orgIDs []identity.ID
	if orgID != nil {
		orgIDs = []identity.ID{*orgID}
	}

	var projectNames []string
	if name != nil {
		projectNames = []string{*name}
	}

	return client.Projects.Search(c, orgIDs, projectNames)
}

func listProjectsByOrgID(ctx *context.Context, client *api.Client, orgIDs []identity.ID) ([]envelope.Project, error) {
	c, client, err := NewAPIClient(ctx, client)
	if err != nil {
		return nil, cli.NewExitError(projectListFailed, -1)
	}

	return client.Projects.Search(c, orgIDs, nil)
}

func listProjectsByOrgName(ctx *context.Context, client *api.Client, orgName string) ([]envelope.Project, error) {
	c, client, err := NewAPIClient(ctx, client)
	if err != nil {
		return nil, cli.NewExitError(projectListFailed, -1)
	}

	// Look up the target org
	var org *envelope.Org
	org, err = client.Orgs.GetByName(c, orgName)
	if err != nil {
		return nil, errs.NewExitError(projectListFailed)
	}
	if org == nil {
		return nil, errs.NewExitError("Org not found.")
	}

	// Pull all projects for the given orgID
	orgIDs := []identity.ID{*org.ID}
	projects, err := listProjectsByOrgID(&c, client, orgIDs)
	if err != nil {
		return nil, errs.NewExitError(projectListFailed)
	}

	return projects, nil
}

const projectCreateFailed = "Could not create project."

func createProjectCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(projectCreateFailed)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, orgName, newOrg, err := SelectCreateOrg(c, client, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError(projectCreateFailed, err)
	}

	var orgID *identity.ID
	if !newOrg {
		if org == nil {
			return errs.NewExitError("Org not found.")
		}
		orgID = org.ID
	}

	args := ctx.Args()
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	label := "Project name"
	autoAccept := name != ""
	name, err = NamePrompt(&label, name, autoAccept)
	if err != nil {
		return handleSelectError(err, projectCreateFailed)
	}

	if newOrg {
		org, err := client.Orgs.Create(c, orgName)
		orgID = org.ID
		if err != nil {
			return errs.NewErrorExitError("Could not create org", err)
		}

		err = generateKeypairsForOrg(c, ctx, client, org.ID, false)
		if err != nil {
			return err
		}

		fmt.Printf("Org %s created.\n\n", orgName)
	}

	_, err = createProjectByName(c, client, orgID, name)
	return err
}

func createProjectByName(c context.Context, client *api.Client, orgID *identity.ID, name string) (*envelope.Project, error) {
	project, err := client.Projects.Create(c, orgID, name)
	if orgID == nil {
		return nil, errs.NewExitError("Org not found")
	}
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return nil, errs.NewExitError("Project already exists")
		}
		return nil, errs.NewErrorExitError(projectCreateFailed, err)
	}
	fmt.Printf("Project %s created.\n", name)
	return project, nil
}

func getProject(ctx context.Context, client *api.Client, orgID *identity.ID, name string) (*envelope.Project, error) {
	orgIDs := []identity.ID{*orgID}
	names := []string{name}
	projects, err := client.Projects.Search(ctx, orgIDs, names)
	if projects == nil || len(projects) < 1 {
		return nil, errs.NewExitError("Project not found.")
	}
	if err != nil {
		return nil, errs.NewErrorExitError("Unable to lookup project.", err)
	}

	return &projects[0], nil
}

// This functions is intended to be used when a command takes an optional project flag.
// In the situation where no project is provided in the flags, getProjectWithPrompt prompts the user
// to select from a list of exisitng project, using SelectExistingProjectPrompt.
// In the situation where a project is provided in the flags, the function returns the associated
// project.Envelope structure.
func getProjectWithPrompt(client *api.Client, c context.Context, org *envelope.Org, projectName string) (*envelope.Project, error) {

	var project *envelope.Project

	if projectName == "" {
		// Retrieve list of available orgs
		projects, err := client.Projects.List(c, org.ID)
		if err != nil {
			return nil, err
		}
		if len(projects) < 1 {
			return nil, errs.NewExitError("No projects found in org " + org.Body.Name)
		}

		// Prompt user to select from list of existing projects
		idx, _, err := SelectExistingProjectPrompt(projects)
		if err != nil {
			return nil, err
		}

		project = &projects[idx]

	} else {
		// Get Project for project name, confirm project exists
		var err error
		project, err = getProject(c, client, org.ID, projectName)
		if err != nil {
			return nil, err
		}
		if project == nil {
			return nil, errs.NewExitError("project " + projectName + " not found.")
		}
	}

	return project, nil
}
