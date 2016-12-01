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

	path, target := identifyTarget(args, recursive)
	if target == "" {
		return errs.NewUsageExitError("Invalid path supplied", ctx)
	}

	// Parse the pathexp so we have a full target path
	cpathExp, err := pathexp.ParsePartial(path)
	if err != nil {
		return err
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
		creds, err := client.Credentials.Search(c, cpathExp.String())
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
		return err
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
func identifyTarget(args []string, recursive bool) (string, string) {
	defined := 0
	path := "/"
	if len(args) < 1 {
		if recursive {
			return path, "secrets"
		}
		return path, "orgs"
	}

	var hasDoubleGlob bool
	var showChildren bool
	path = args[0]
	if path != "/" {
		// If path ends with slash, we're looking inside
		if path[len(path)-1:] == "/" {
			showChildren = true
			path = path[:len(path)-1]
		}
		segments := strings.Split(path, "/")
		defined = len(segments) - 1
		if !showChildren {
			defined--
		}
		for _, s := range segments {
			if s == "**" {
				hasDoubleGlob = true
			}
		}
	}

	target := ""
	if len(targetMap) > defined {
		target = targetMap[defined]
	}
	if recursive || hasDoubleGlob {
		return path, "secrets"
	}
	return path, target
}

// retrieve the projecttree for non-recursive operations
func projectTreeForOrg(c context.Context, client *api.Client, cpathExp *pathexp.PathExp) (*api.ProjectTreeSegment, error) {
	org, err := client.Orgs.GetByName(c, cpathExp.Org.String())
	if err != nil {
		return nil, err
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
