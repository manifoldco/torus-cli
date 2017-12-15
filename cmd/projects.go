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
			return errs.NewExitError("Failed to retrieve orgs list.")
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

	projects, err := client.Projects.List(c, org.ID)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve projects list.", err)
	}

	fmt.Println("")
	fmt.Printf("  %s\n", ui.Bold("Projects"))
	for _, project := range projects {
		fmt.Printf("  %s\n", project.Body.Name)
	}
	fmt.Println("")

	count := strconv.Itoa(len(projects))
	countStr := "Org " + orgName + " has (" + count + ") projects."
	fmt.Println(countStr)

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
		return nil, errs.NewErrorExitError("Project not found.", err)
	}
	if err != nil {
		return nil, errs.NewErrorExitError("Unable to lookup project.", err)
	}

	return &projects[0], nil
}
