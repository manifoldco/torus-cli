package cmd

import (
	"fmt"
	"context"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/pathexp"
)

func init() {
	list := cli.Command{
		Name:      "list",
		ArgsUsage: "",
		Usage:     "List your org's project structure and its secrets.",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			envFlag("Use this environment.", false),
			serviceFlag("Use this service.", "", false),
			nameFlag("Find secrets with this name."),
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Lists the sources of the secrets (shortcut for --format verbose)",
			},
		},
		Action: chain(
			//ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			//setUserEnv, checkRequiredFlags, listCmd,
			ensureDaemon, ensureSession, checkRequiredFlags, listCmd,
		),
	}
	Cmds = append(Cmds, list)
}

type ServiceSecretList struct{
	service string
	secrets []string
}

func listCmd(ctx *cli.Context) error {

	verbose := false
	if ctx.Bool("verbose"){
		verbose = true
	}

	envToServiceSecret := make(map[string][]ServiceSecretList)

	orgName := ctx.String("org")
	projectName := ctx.String("project")
	envFilter := ctx.String("environment")
	serviceFilter := ctx.String("service")
	secretFilter := ctx.String("name")

	projectPath := "/" + orgName + "/" + projectName + "/"

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Get Org
	org, err := getOrg(c, client, orgName)
	if err != nil {
		return err
	}

	// Get Project
	project, err := getProject(c, client, org.ID, projectName)
	if err != nil {
		return err
	}

	// Get environment names
	envNames := []string{}
	if envFilter != "" {
		envNames = append(envNames, envFilter)
	} else {		
		environments, err := listEnvs(&c, client, org.ID, project.ID, nil)
		if err != nil {
			return err
		}
		for _, e := range environments{
			envNames = append(envNames, e.Body.Name)
		}
	}

	// Get service names
	serviceNames := []string{}
	if serviceFilter != "" {
		serviceNames = append(serviceNames, serviceFilter)
	} else {
		services, err := listServices(&c, client, org.ID, project.ID, nil)
		if err != nil {
			return err
		}
		for _, s := range services{
			serviceNames = append(serviceNames, s.Body.Name)
		}
	}

	// Retrieve secrets under each env/service combo
	for _, e := range envNames{
		for _, s := range serviceNames{
			pexp, err := pathexp.ParsePartial(projectPath + e + "/" + s + "/" + "*/*")
			credentials, err := client.Credentials.Search(c, pexp.String())
			if err != nil {
				return err
			}
			if len(credentials) == 0 {
				continue
			}
			var credNames []string
			for _, c := range credentials{
				credName := (*c.Body).GetName()
				if secretFilter != "" && (*c.Body).GetName() != secretFilter {
					continue
				}
				if verbose == true{
					credName = (*c.Body).GetPathExp().String() + "/" + credName
				}
				credNames = append(credNames, credName)
			}
			if len(credNames) > 0 {
				envToServiceSecret[e] = append(envToServiceSecret[e], ServiceSecretList {
					service: s,
					secrets: credNames,
				})
			}
		}
	}

	fmt.Println(projectPath)
	for k, v := range envToServiceSecret{
		fmt.Println("\t", k)
		for _, s := range v{
			fmt.Println("\t\t", s.service)
			for _, secret := range s.secrets{
				fmt.Println("\t\t\t", secret)
			}
		}
	}

	return nil
}
