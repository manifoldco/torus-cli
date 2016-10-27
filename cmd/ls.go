package cmd

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
)

func init() {
	ls := cli.Command{
		Name:      "ls",
		ArgsUsage: "[cpath]",
		Usage:     "Explore all objects your account has access to",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			orgFlag("Use this organization.", false),
			projectFlag("Use this project.", false),
			newSlicePlaceholder("environment, e", "ENV", "Use this environment.",
				"*", "TORUS_ENVIRONMENT", false),
			newSlicePlaceholder("service, s", "SERVICE", "Use this service.",
				"default", "TORUS_SERVICE", false),
			newSlicePlaceholder("user, u", "USER", "Use this user (identity).",
				"*", "TORUS_USER", false),
			newSlicePlaceholder("instance, i", "INSTANCE", "Use this instance.",
				"*", "TORUS_INSTANCE", false),
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			checkRequiredFlags, listObjects,
		),
	}
	Cmds = append(Cmds, ls)
}

func listObjects(ctx *cli.Context) error {
	args := ctx.Args()

	var useFlags bool

	// Try to be helpful and detect when a wildcard isn't quoted
	if len(args) > 1 {
		return errs.NewUsageExitError("Invalid path supplied.\n"+
			"Note: arguments containing wildcards must be wrapped in qutoes.", ctx)
	}

	var cpathObj *pathexp.PathExp
	var count int
	var err error

	hasPath := len(args) > 0
	emptyOrg := ctx.String("org") == ""
	emptyProj := ctx.String("project") == ""
	emptyOrgOrProject := len(args) == 0 && (emptyOrg || emptyProj)

	if hasPath || emptyOrgOrProject {
		var path string
		if hasPath {
			// User has supplied a path
			path = args[0]
		} else if emptyProj {
			// Context enabled, no project supplied
			path = "/" + ctx.String("org")
		} else {
			// Context enabled, no org supplied
			path = "/"
		}

		// Construct the target from single path supplied
		obj, segmentCount, err := pathexp.NewPartialFromPath(path)
		cpathObj = obj
		count = *segmentCount
		if err != nil {
			return err
		}
	} else {
		// Construct the target path through flags
		obj, _, err := pathexp.NewPartial(ctx.String("org"), ctx.String("project"),
			ctx.StringSlice("environment"), ctx.StringSlice("service"),
			ctx.StringSlice("user"), ctx.StringSlice("instance"),
		)
		cpathObj = obj
		count = 6
		if err != nil {
			return err
		}
		useFlags = true
	}

	// When using flag composition, require org and project at minimum
	if useFlags && count < 1 {
		return errs.NewUsageExitError("Org must be supplied", ctx)
	}
	if useFlags && count < 2 {
		return errs.NewUsageExitError("Project must be supplied", ctx)
	}
	if count > 3 && cpathObj.Org() == "*" {
		return errs.NewUsageExitError("Org must be supplied to view secrets", ctx)
	}
	if count > 3 && cpathObj.Project() == "*" {
		return errs.NewUsageExitError("Project must be supplied to view secrets", ctx)
	}

	var paths []string
	var pathErr error

	showing := "Listing objects for path"
	switch count {
	case 0:
		showing = "Listing orgs:"
		paths, err = orgPaths()
		if err != nil {
			pathErr = errs.NewExitError("Failed to list orgs.")
		}
	case 1:
		partial := "/" + cpathObj.Org()
		showing = "Listing projects within path: " + partial
		paths, err = projectPaths(cpathObj)
		if err != nil {
			pathErr = errs.NewExitError("Failed to list projects.")
		}
	case 2:
		partial := "/" + cpathObj.Org() + "/" + cpathObj.Project()
		showing = "Listing environments within path: " + partial
		paths, err = envPaths(cpathObj)
		if err != nil {
			pathErr = errs.NewExitError("Failed to list environments.")
		}
	case 3:
		partial := "/" + cpathObj.Org() + "/" + cpathObj.Project() + "/*"
		showing = "Listing services within path: " + partial
		paths, err = servicePaths(cpathObj)
		if err != nil {
			pathErr = errs.NewExitError("Failed to list services.")
		}
	default:
		showing = "Listing secrets within path: " + cpathObj.String()
		paths, err = secretPaths(cpathObj)
		if err != nil {
			pathErr = errs.NewExitError("Failed to list secrets.")
		}
	}

	// Print the path and object type of the output
	fmt.Println(showing)

	// Must replace Envs with * when only 3 segments supplied
	if count == 3 && cpathObj.Envs() != "*" {
		fmt.Println("Note: environment exchanged for *")
		fmt.Println("")
	} else {
		fmt.Println("")
	}

	if pathErr != nil {
		return pathErr
	}
	for _, path := range paths {
		fmt.Println(path)
	}

	return nil
}

// generate all paths for /
func orgPaths() ([]string, error) {
	var paths []string

	orgs, _, err := orgsList()
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		paths = append(paths, "/"+org.Body.Name)
	}

	return paths, nil
}

// generate all paths for /${org}
func projectPaths(cpathObj *pathexp.PathExp) ([]string, error) {
	var paths []string

	orgs, _, err := orgsList()
	if err != nil {
		return nil, err
	}

	orgIDs, _, orgMapID, err := filterOrgs(orgs, cpathObj)
	if err != nil {
		return nil, err
	}

	projects, err := listProjectsByOrgID(nil, nil, orgIDs)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		pOrgName := orgMapID[*project.Body.OrgID]
		paths = append(paths, fmt.Sprintf("/%s/%s", pOrgName, project.Body.Name))
	}

	return paths, nil
}

// generate all paths for /${org}/${project}
func envPaths(cpathObj *pathexp.PathExp) ([]string, error) {
	var paths []string
	c, client, err := NewAPIClient(nil, nil)

	orgs, _, err := orgsList()
	if err != nil {
		return nil, err
	}

	orgIDs, _, orgMapID, err := filterOrgs(orgs, cpathObj)
	if err != nil {
		return nil, err
	}

	projects, err := listProjectsByOrgID(&c, client, orgIDs)
	if err != nil {
		return nil, err
	}

	projectIDs, projMapID, err := filterProjects(projects, cpathObj)
	if err != nil {
		return nil, err
	}

	envs, err := listEnvsByProjectID(&c, client, projectIDs)
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		eOrgName := orgMapID[*env.Body.OrgID]
		eProjName := projMapID[*env.Body.ProjectID]
		paths = append(paths, fmt.Sprintf("/%s/%s/%s", eOrgName, eProjName, env.Body.Name))
	}

	return paths, nil
}

// generate all paths for /${org}/${project}/*
func servicePaths(cpathObj *pathexp.PathExp) ([]string, error) {
	var paths []string
	c, client, err := NewAPIClient(nil, nil)

	orgs, _, err := orgsList()
	if err != nil {
		return nil, err
	}

	orgIDs, _, orgMapID, err := filterOrgs(orgs, cpathObj)
	if err != nil {
		return nil, err
	}

	projects, err := listProjectsByOrgID(&c, client, orgIDs)
	if err != nil {
		return nil, err
	}

	projectIDs, projMapID, err := filterProjects(projects, cpathObj)
	if err != nil {
		return nil, err
	}

	services, err := listServicesByProjectID(&c, client, projectIDs)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		sOrgName := orgMapID[*service.Body.OrgID]
		sProjName := projMapID[*service.Body.ProjectID]
		paths = append(paths, fmt.Sprintf("/%s/%s/*/%s", sOrgName, sProjName, service.Body.Name))
	}

	return paths, nil
}

func secretPaths(cpathObj *pathexp.PathExp) ([]string, error) {
	var paths []string
	c, client, err := NewAPIClient(nil, nil)

	creds, err := client.Credentials.Search(c, cpathObj.String())
	if err != nil {
		return nil, err
	}

	cset := credentialSet{}
	for _, c := range creds {
		cset.Add(c)
	}

	for _, cred := range cset {
		body := *cred.Body
		paths = append(paths, fmt.Sprintf("%s/%s", body.GetPathExp(), body.GetName()))
	}
	sort.Strings(paths)

	return paths, nil
}

func filterProjects(projects []api.ProjectResult, pe *pathexp.PathExp) ([]*identity.ID, map[identity.ID]string, error) {
	projMapID := make(map[identity.ID]string, len(projects))
	var projectIDs []*identity.ID
	for _, project := range projects {
		projMapID[*project.ID] = project.Body.Name
		if pe.Project() == "*" {
			projectIDs = append(projectIDs, project.ID)
		} else if project.Body.Name == pe.Project() {
			projectIDs = append(projectIDs, project.ID)
		}
	}
	if len(projectIDs) < 1 {
		return nil, nil, errs.NewExitError("Project not found.")
	}
	return projectIDs, projMapID, nil
}

func filterOrgs(orgs []api.OrgResult, pe *pathexp.PathExp) ([]*identity.ID, map[string]identity.ID, map[identity.ID]string, error) {
	orgMapName := make(map[string]identity.ID, len(orgs))
	orgMapID := make(map[identity.ID]string, len(orgs))

	var orgIDs []*identity.ID
	for _, org := range orgs {
		orgMapName[org.Body.Name] = *org.ID
		orgMapID[*org.ID] = org.Body.Name
		if pe.Org() == "*" || org.Body.Name == pe.Org() {
			orgIDs = append(orgIDs, org.ID)
		}
	}
	if len(orgIDs) < 1 {
		return nil, nil, nil, errs.NewExitError("Org not found.")
	}

	return orgIDs, orgMapName, orgMapID, nil
}
