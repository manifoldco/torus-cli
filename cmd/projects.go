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
					OrgFlag("List projects in an organization", true),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, listProjects,
				),
			},
		},
	}
	Cmds = append(Cmds, projects)
}

const projectListFailed = "Coult not list projects, please try again."

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
