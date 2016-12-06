package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/pathexp"
)

var targetMap = []string{"orgs", "projects", "envs", "services"}

func init() {
	ls := cli.Command{
		Name:      "ls",
		ArgsUsage: "<path>",
		Usage:     "Explore all objects your account has access to",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			recursiveFlag(),
			formatFlag("simple", "Format used to display data (simple, verbose)"),
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Lists the types of resources and source path (shortcut for --format verbose)",
			},
			orgFlag("Use this organization.", false),
			projectFlag("Use this project.", false),
			newSlicePlaceholder("environment, e", "ENV", "Use this environment.",
				"*", "TORUS_ENVIRONMENT", false),
			newSlicePlaceholder("service, s", "SERVICE", "Use this service.",
				"default", "TORUS_SERVICE", false),
			newSlicePlaceholder("user, u", "USER", "Use this user (identity).",
				"*", "TORUS_USER", false),
			newSlicePlaceholder("machine, m", "MACHINE", "Use this machine.",
				"*", "TORUS_MACHINE", false),
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
	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError("Failed to retrieve objects")
	}

	client := api.NewClient(cfg)
	c := context.Background()

	args := ctx.Args()
	recursive := ctx.Bool("recursive")

	format := ctx.String("format")
	if ctx.Bool("verbose") {
		format = "verbose"
	}
	if format != "verbose" && format != "simple" {
		return errs.NewUsageExitError("Invalid format", ctx)
	}

	// Try to be helpful and detect when a wildcard isn't quoted
	if len(args) > 1 {
		return errs.NewUsageExitError("Invalid path supplied.\n"+
			"Note: arguments containing wildcards must be wrapped in quotes.", ctx)
	}

	cpathExp, target, err := identifyTarget(args, recursive)
	if err != nil || cpathExp == nil {
		return errs.NewUsageExitError("Invalid path supplied", ctx)
	}

	var orgName string
	var projectTree api.ProjectTreeSegment
	var projectMap map[string]*api.ProjectResult
	if target != "orgs" {
		tree, err := projectTreeForOrg(c, client, cpathExp)
		if err != nil {
			return err
		}
		projectTree = *tree
		orgName = projectTree.Org.Body.Name
		projectMap = matchingProjects(cpathExp, projectTree)
	}

	// Pull list of paths for the target object
	var paths []string
	var pathsErr error
	switch target {
	case "orgs":
		orgs, _, err := orgsList()
		if err != nil {
			pathsErr = err
			break
		}
		for _, o := range orgs {
			if cpathExp.Org.Contains(o.Body.Name) {
				paths = append(paths, fmt.Sprintf("/%s", o.Body.Name))
			}
		}
	case "projects":
		for _, p := range projectMap {
			paths = append(paths, fmt.Sprintf("/%s/%s", orgName, p.Body.Name))
		}
	case "envs":
		for _, e := range projectTree.Envs {
			if p, ok := projectMap[e.Body.ProjectID.String()]; ok {
				if cpathExp.Envs.Contains(e.Body.Name) {
					paths = append(paths, fmt.Sprintf("/%s/%s/%s", orgName, p.Body.Name, e.Body.Name))
				}
			}
		}
	case "services":
		for _, s := range projectTree.Services {
			if p, ok := projectMap[s.Body.ProjectID.String()]; ok {
				if cpathExp.Services.Contains(s.Body.Name) {
					paths = append(paths, fmt.Sprintf("/%s/%s/%s/%s", orgName, p.Body.Name, "*", s.Body.Name))
				}
			}
		}
	case "secrets":
		segments := strings.Split(args[0], "/")
		targetName := segments[len(segments)-1:][0]
		pexp, err := pathexp.Parse(cpathExp.String())
		if err != nil {
			pathsErr = err
			break
		}
		creds, err := client.Credentials.Search(c, pexp.String())
		if err != nil {
			pathsErr = err
			break
		}
		cset := credentialSet{}
		for _, c := range creds {
			cset.Add(c)
		}
		for _, cred := range cset {
			body := *cred.Body
			paths = append(paths, fmt.Sprintf("%s/%s", body.GetPathExp(), body.GetName()))
		}
	default:
		pathsErr = errs.NewUsageExitError("Unknown path supplied", ctx)
	}
	if pathsErr != nil {
		return pathsErr
	}

	// Final output of paths
	if format == "verbose" {
		fmt.Println(strings.ToUpper(target) + "\n")
	}
	sort.Strings(paths)
	for _, p := range paths {
		fmt.Println(p)
	}

	return nil
}

// return a map of the projects which match the supplied pathexp
func matchingProjects(pexp *pathexp.PathExp, tree api.ProjectTreeSegment) map[string]*api.ProjectResult {
	projectMap := make(map[string]*api.ProjectResult)
	for _, p := range tree.Projects {
		if pexp.Project.Contains(p.Body.Name) {
			projectMap[p.ID.String()] = p
		}
	}
	return projectMap
}

// identify whether we want to list children or matching resources
func identifyTarget(args []string, recursive bool) (*pathexp.PathExp, string, error) {
	var defined int
	var path string
	var target string
	var hasDoubleGlob bool
	var showChildren bool

	// Default to org lookup
	if len(args) != 1 {
		path = "/"
		target = "orgs"
	} else {
		path = args[0]
		if len(path) == 0 {
			return nil, "", errs.NewExitError("Invalid path supplied")
		}
	}

	if path != "/" {
		// Paths must begin with slash
		if path[:1] != "/" {
			return nil, "", errs.NewExitError("path must start with /")
		}
		// Identify if path ends with slash for child lookup
		if path[len(path)-1:] == "/" {
			showChildren = true
			path = path[:len(path)-1]
		}
		segments := strings.Split(path, "/")
		defined = len(segments) - 1
		if !showChildren {
			defined--
		}
		// Identify if double glob was used in any segment
		for _, s := range segments {
			if s == "**" {
				hasDoubleGlob = true
			}
		}
	}

	// Pull target from map
	if len(targetMap) > defined {
		target = targetMap[defined]
	}
	if recursive || hasDoubleGlob {
		target = "secrets"
	}

	// Parse the partial pathexp to be used with matching
	pexp, err := pathexp.ParsePartial(path)
	if err != nil {
		return nil, "", err
	}

	// valid org required to list objects in the tree
	if target != "orgs" && !pathexp.ValidSlug(pexp.Org.String()) {
		return nil, "", errs.NewExitError("Invalid path supplied")
	}
	// valid project must be present for non-project targets
	if target != "orgs" && target != "projects" && !pathexp.ValidSlug(pexp.Project.String()) {
		return nil, "", errs.NewExitError("Invalid path supplied")
	}

	return pexp, target, nil
}

// retrieve the projecttree for non-recursive operations
func projectTreeForOrg(c context.Context, client *api.Client, cpathExp *pathexp.PathExp) (*api.ProjectTreeSegment, error) {
	org, err := client.Orgs.GetByName(c, cpathExp.Org.String())
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errs.NewExitError("Org not found")
	}

	projectTree, err := client.Projects.GetTree(c, org.ID)
	if err != nil {
		return nil, errs.NewErrorExitError("Could not retrieve project information", err)
	}
	if len(projectTree) > 0 {
		return &projectTree[0], nil
	}

	return nil, errs.NewExitError("Could not retrieve project information")
}
