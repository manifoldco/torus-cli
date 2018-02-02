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
					orgFlag("org to show services for", false),
					projectFlag("project to show services for", false),
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

	project, _, _, err := selectProject(c, client, org, ctx.String("project"), false)
	if err != nil {
		return err
	}

	// Retrieve services for targeted org and project
	services, err := listServices(&c, client, org.ID, project.ID, nil, nil)
	if err != nil {
		return errs.NewErrorExitError(serviceListFailed, err)
	}

	// Build output of projects/envs
	fmt.Println("")
	fmt.Printf("%s\n", ui.BoldString("Services"))
	for _, s := range services {
		fmt.Printf("%s\n", s.Body.Name)
	}

	fmt.Printf("\nProject /%s/%s has (%s) service%s\n", org.Body.Name, project.Body.Name, ui.FaintString(strconv.Itoa(len(services))), plural(len(services)))

	return nil
}

func listServices(ctx *context.Context, client *api.Client, orgID, projID *identity.ID, teamIDs []identity.ID, name *string) ([]envelope.Service, error) {
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

	return client.Services.List(c, orgIDs, projectIDs, nil, names)
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

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	project, _, _, err := selectProject(c, client, org, ctx.String("project"), false)
	if err != nil {
		return err
	}

	serviceName, err = prompts.ServiceName(serviceName, true)
	if err != nil {
		return err
	}

	// Create our new service
	fmt.Println("")
	err = client.Services.Create(c, org.ID, project.ID, serviceName)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Service already exists")
		}
		return errs.NewErrorExitError(serviceCreateFailed, err)
	}

	fmt.Printf("Service %s created.\n", serviceName)
	return nil
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
