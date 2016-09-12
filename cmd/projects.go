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
		return cli.NewExitError(projectListFailed, -1)
	}
	if org == nil {
		return cli.NewExitError("Org not found.", -1)
	}

	// Pull all projects for the given orgID
	projects, err := client.Projects.List(c, org.ID, nil)
	if err != nil {
		return cli.NewExitError(projectListFailed, -1)
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
		return cli.NewExitError(projectCreateFailed, -1)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, orgName, newOrg, err := SelectCreateOrg(client, c, ctx.String("org"))

	var orgID *identity.ID
	if !newOrg {
		if org == nil {
			return cli.NewExitError("Org not found", -1)
		}
		orgID = org.ID
	}

	args := ctx.Args()
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	label := "Project name"
	if name == "" {
		name, err = NamePrompt(&label, "")
		if err != nil {
			if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
				return err
			}
			fmt.Println("")
			return cli.NewExitError(projectCreateFailed, -1)
		}
	} else {
		fmt.Println(promptui.SuccessfulValue(label, name))
	}

	if newOrg {
		org, err := client.Orgs.Create(c, orgName)
		orgID = org.ID
		if err != nil {
			return cli.NewExitError("Could not create org: "+err.Error(), -1)
		}

		err = generateKeypairsForOrg(ctx, c, client, org.ID, false)
		if err != nil {
			return err
		}

		fmt.Printf("Org %s created.\n\n", orgName)
	}

	_, err = createProjectByName(c, client, orgID, name)
	if err != nil {
		return err
	}

	return nil
}

func createProjectByName(c context.Context, client *api.Client, orgID *identity.ID, name string) (*api.ProjectResult, error) {
	project, err := client.Projects.Create(c, orgID, name)
	if orgID == nil {
		return nil, cli.NewExitError("Unknown org", -1)
	}
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return nil, cli.NewExitError("Project already exists", -1)
		}
		fmt.Println(err)
		return nil, cli.NewExitError(projectCreateFailed, -1)
	}
	fmt.Printf("Project %s created.\n", name)
	return project, nil
}
