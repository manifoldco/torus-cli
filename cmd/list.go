package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/juju/ansiterm"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/ui"
)

func init() {
	list := cli.Command{
		Name:      "list",
		ArgsUsage: "",
		Usage:     "List allows you to list and filter all the secrets that you can access inside a project.",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			orgFlag("Use this organization.", false),
			projectFlag("Use this project.", false),
			envSliceFlag("Use this environment.", false),
			serviceSliceFlag("Use this service.", "", false),
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Display the full credential path of each secret.",
			},
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, checkRequiredFlags, listCmd,
		),
	}
	Cmds = append(Cmds, list)
}

type serviceCredentialMap map[string]credentialSet
type credentialTree map[string]serviceCredentialMap

func listCmd(ctx *cli.Context) error {
	verbose := ctx.Bool("verbose")

	args := ctx.Args()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	tree := make(credentialTree)

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	project, _, _, err := selectProject(c, client, org, ctx.String("project"), false)
	if err != nil {
		return err
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
	// build the PathExp later.
	instanceFilters := []string{"*"}
	idenityFilters := []string{"*"}

	// Create a PathExp based on flags. This is the search space that
	// will be used to retrieve credentials.
	filterPathExp, err := pathexp.New(
		org.Body.Name,
		project.Body.Name,
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

	go func() {
		// Get environments
		environments, eErr = listEnvs(&c, client, org.ID, project.ID, nil)
		getEnvsServicesCreds.Done()
	}()

	go func() {
		// Get services
		services, sErr = listServices(&c, client, org.ID, project.ID, nil)
		getEnvsServicesCreds.Done()
	}()

	go func() {
		// Get credentials
		credentials, cErr = client.Credentials.Search(c, filterPathExp.String(), nil)
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
		if filterPathExp.Envs.Contains(e.Body.Name) {
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

	// Create credentialsTree for verbose mode
	// In verbose mode, ALL paths are displayed,
	// whether they contain credentials or not.
	// This will be filled in the following section.
	if verbose {
		for _, eName := range filteredEnvNames {
			tree[eName] = make(serviceCredentialMap)
			for _, sName := range filteredServiceNames {
				tree[eName][sName] = make(credentialSet)
			}
		}
	}

	// Check for each env and service in the filtered list, add
	// any credentials along that path to that env/service branch
	// of the credentialsTree
	credCount := 0
	projectPath := "/" + org.Body.Name + "/" + project.Body.Name + "/"
	for _, e := range filteredEnvNames {
		for _, s := range filteredServiceNames {
			builtPathExp, err := pathexp.Parse(projectPath + e + "/" + s + "/*/*")
			if err != nil {
				return errs.NewErrorExitError("Failed to parse: "+projectPath+e+"/"+s+"/*/*", err)
			}
			for _, cred := range credentials {
				body := *cred.Body
				if len(args) > 0 && !isSecretNameInList(body.GetName(), args) {
					continue
				}
				credPathExp := body.GetPathExp()
				// If cred not contained in any builtPathExps, it is not
				// within the search space specified by the flags.
				if !credPathExp.Contains(builtPathExp) {
					continue
				}
				// "Add" is defined in 'credential_set.go'. This
				// handles the case where a secret is redefined in
				// overlapping spaces.
				if tree[e] == nil {
					tree[e] = make(serviceCredentialMap)
				}
				if tree[e][s] == nil {
					tree[e][s] = make(credentialSet)
				}
				tree[e][s].Add(cred)
			}
		}
	}

	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 0, 0, 0, ' ', 0)
	for e := range tree {
		fmt.Fprintf(w, fmt.Sprintf("%s\t\t\t\t\n", ui.BoldString(e)+"/"))
		for s := range tree[e] {
			fmt.Fprintf(w, "\t%s\t\t\t\n", ui.BoldString(s)+"/")
			if len(tree[e][s]) == 0 {
				if verbose {
					fmt.Fprintf(w, "\t\t%s\t\t\n", ui.FaintString("[empty]"))
				}
				continue
			}
			for c, cred := range tree[e][s] {
				credCount++
				if verbose {
					credPath := (*cred.Body).GetPathExp().String() + "/"
					fmt.Fprintf(w, "\t\t%s\t(%s)\t\n", c, ui.FaintString(credPath+c))
				} else {
					fmt.Fprintf(w, "\t\t%s\t\t\n", c)
				}
			}
		}
	}
	w.Flush()

	fmt.Printf("\n(%s) secrets found\n", ui.FaintString(strconv.Itoa(credCount)))

	return nil
}

func isSecretNameInList(secret string, list []string) bool {
	for _, s := range list {
		if s == secret {
			return true
		}
	}
	return false
}
