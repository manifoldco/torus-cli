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
	"github.com/manifoldco/torus-cli/prompts"
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

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	projects, err := client.Projects.List(c, org.ID)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve projects list.", err)
	}

	fmt.Println("")
	fmt.Printf("%s\n", ui.BoldString("Projects"))
	for _, project := range projects {
		fmt.Printf("%s\n", project.Body.Name)
	}

	fmt.Printf("\nOrg %s has (%s) project%s\n", org.Body.Name,
		ui.FaintString(strconv.Itoa(len(projects))), plural(len(projects)))

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

const projectCreateFailed = "Could not create project."

func createProjectCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(projectCreateFailed)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	args := ctx.Args()
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	name, err = prompts.ProjectName(name, true)
	if err != nil {
		return err
	}

	_, err = createProjectByName(c, client, org.ID, name)
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
