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
	"github.com/arigatomachine/cli/errs"
	"github.com/arigatomachine/cli/identity"
)

func init() {
	projects := cli.Command{
		Name:     "projects",
		Usage:    "View and manipulate projects within an organization",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List services for an organization",
				Flags: []cli.Flag{
					orgFlag("List projects in an organization", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, listProjects,
				),
			},
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
		},
	}
	Cmds = append(Cmds, projects)
}

const projectListFailed = "Could not list projects, please try again."

func listProjects(ctx *cli.Context) error {
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
		return errs.NewExitError(projectListFailed)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	// Pull all projects for the given orgID
	projects, err := client.Projects.List(c, org.ID, nil)
	if err != nil {
		return errs.NewExitError(projectListFailed)
	}

	fmt.Println("")
	count := strconv.Itoa(len(projects))
	title := org.Body.Name + " org (" + count + ")"
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", utf8.RuneCountInString(title)))
	for _, project := range projects {
		fmt.Println(project.Body.Name)
	}
	fmt.Println("")

	return nil
}

const projectCreateFailed = "Could not create project. Please try again."

func createProjectCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(projectCreateFailed)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, orgName, newOrg, err := SelectCreateOrg(c, client, ctx.String("org"))

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

func createProjectByName(c context.Context, client *api.Client, orgID *identity.ID, name string) (*api.ProjectResult, error) {
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
