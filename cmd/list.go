package cmd

import (
	"fmt"
	"context"
	//"strings"
	"sync"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/envelope"
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
			envSliceFlag("Use this environment.", false),
			serviceSliceFlag("Use this service.", "", false),
			nameFlag("Find secrets with this name."),
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Lists the sources of the secrets (shortcut for --format verbose)",
			},
		},
		Action: chain(
			ensureDaemon, ensureSession, checkRequiredFlags, listCmd,
		),
	}
	Cmds = append(Cmds, list)
}

func listCmd(ctx *cli.Context) error {

	verbose := ctx.Bool("verbose")

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// TODO - Make this a type
	credentialsTree := make(map[string]map[string]credentialSet)

	// Retrieve org and project flag values
	orgName := ctx.String("org")
	projectName := ctx.String("project")

	// Get Org for org name, confirm org exists
	org, err := getOrg(c, client, orgName)
	if err != nil {
		return errs.NewErrorExitError("Failed to get org info for " +
			orgName, err)
	}

	// Get Project for project name, confirm project exists
	project, err := getProject(c, client, org.ID, projectName)
	if err != nil {
		return errs.NewErrorExitError("Failed to get project info for " +
			orgName + "/" + projectName, err)
	}

	// Retrieve environment flag values
	// If no values were set, use a full glob
	envFilters := ctx.StringSlice("environment")
	if len(envFilters) == 0 {
		envFilters = append(envFilters, "*")
	}

	// Retrieve service flag values
	// If no values were set, use a full glob
	serviceFilters := ctx.StringSlice("service")
	if len(serviceFilters) == 0 {
		serviceFilters = append(serviceFilters, "*")
	}

	// The following two slices are placeholders necessary to
	// build the PathExp later. Check that this is true, if I pass
	// empty slices, will it replace those with wildcards?
	instanceFilters := []string{"*"}
	idenityFilters := []string{"*"}

	// Create a PathExp based on flags. This is the search space that
	// will be used to retrieve credentials.
	filterPathExp, err := pathexp.New(
		orgName,
		projectName,
		envFilters,
		serviceFilters,
		instanceFilters,
		idenityFilters)
	if err != nil {
		return errs.NewErrorExitError("Failed to create path for specified flags.", err)
	}

	// Retrieve envs, services and credentials in parallel
	var getEnvsServicesCreds sync.WaitGroup
	getEnvsServicesCreds.Add(3)

	var environments []envelope.Environment
	var services []envelope.Service
	var credentials []apitypes.CredentialEnvelope
	var eErr, sErr, cErr error

	go func(){
		// Get environments
		environments, eErr = listEnvs(&c, client, org.ID, project.ID, nil)
		getEnvsServicesCreds.Done()
	}()

	go func(){
		// Get services
		services, sErr = listServices(&c, client, org.ID, project.ID, nil)
		getEnvsServicesCreds.Done()
	}()

	go func() {
		// Get credentials
		credentials, cErr = client.Credentials.Search(c, filterPathExp.String())
		getEnvsServicesCreds.Done()
	}()

	getEnvsServicesCreds.Wait()

	if cErr != nil {
		return errs.NewErrorExitError("Could not retrieve credentials.", cErr)
	}

	if eErr != nil {
		return errs.NewErrorExitError("Could not retrieve environments.", eErr)
	}

	if sErr != nil {
		return errs.NewErrorExitError("Could not retrieve services.", sErr)
	}

	filteredEnvNames := []string{}
	filteredServiceNames := []string{}

	// Filter out the retrieved environments based on the
	// search space provided in filterPathExp. If no flags
	// were set, all environments will pass the following test.
	for _, e := range environments {
		if filterPathExp.Envs.Contains(e.Body.Name){
			filteredEnvNames = append(filteredEnvNames, e.Body.Name)
		}
	}

	// Filter out the retrieved services based on the
	// search space provided in filterPathExp. If no flags
	// were set, all services will pass the following test.
	for _, s := range services {
		if filterPathExp.Services.Contains(s.Body.Name) {
			filteredServiceNames = append(filteredServiceNames, s.Body.Name)
		}
	}

	// Create credentialsTree
	// This will be filled in the following section.
	for _, eName := range filteredEnvNames {
		credentialsTree[eName] = make(map[string]credentialSet)
		for _, sName := range filteredServiceNames {
			credentialsTree[eName][sName] = make(credentialSet)
		}
	}

	// Check for each env and service in the filtered list, add
	// any credentials along that path to that env/service branch
	// of the credentialsTree
	projectPath := "/" + orgName + "/" + projectName + "/"
	for _, e := range filteredEnvNames{
		for _, s := range filteredServiceNames{
			for _, cred := range credentials{
				credPathExp := (*cred.Body).GetPathExp()
				builtPathExp, err := pathexp.Parse(projectPath + e + "/" + s + "/*/*")
				if err != nil {
					return errs.NewErrorExitError("Failed to parse: " + projectPath + e + "/" + s + "/*/*", err)
				}
				// If cred not contained in any builtPathExps, it is not
				// within the search space specified by the flags.
				if credPathExp.Contains(builtPathExp) == false {
					continue
				}
				// "Add" is defined in 'credential_set.go'. This
				// handles the case where a secret is redefined in
				// overlapping spaces.
				credentialsTree[e][s].Add(cred)
			}
		}
	}

	// Print credentialsTree
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 5, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\n", projectPath)
	for e, _ := range credentialsTree{
		fmt.Fprintf(w, "\t%s\t\n", e + "/")
		for s, _ := range credentialsTree[e]{
			fmt.Fprintf(w, "\t\t%s\t\n", s + "/")
			if len(credentialsTree[e][s]) == 0 && verbose{
				fmt.Fprintf(w, "\t\t\t[empty]\n")
			} else {
				for c, cred := range credentialsTree[e][s]{
					if verbose{
						credPath := (*cred.Body).GetPathExp().String() + "/"
						fmt.Fprintf(w, "\t\t\t%s\t%s\t\n", c, credPath + c)
					} else {
						fmt.Fprintf(w, "\t\t\t%s\t\t\n", c)
					}
				}
			}
		}
	}
	w.Flush()

	return nil
}
