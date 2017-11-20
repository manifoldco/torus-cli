package cmd

import (
	"fmt"
	"context"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/registry"
)

type TreeNode struct {
	doDisplay bool
	path string
	value string
	secrets []apitypes.CredentialEnvelope //??? Is this stupid?
	parent *TreeNode
	children []*TreeNode
}



func init() {
	list := cli.Command{
		Name:      "list",
		ArgsUsage: "",
		Usage:     "List your org's project structure and its secrets.",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Lists the sources of the secrets (shortcut for --format verbose)",
			},
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, checkRequiredFlags, listCmd,
		),
	}
	Cmds = append(Cmds, list)
}

func testCommand(ctx *cli.Context) error {
	fmt.Println("YOU JUST RAN TOOOOOOOORUS LIST!")
	return nil
}

func treeTest(ctx *cli.Context) error {

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	session, err := client.Session.Who(c)
	if err != nil {
		return err
	}

	identity, err := deriveIdentity(ctx, session)
	if err != nil {
		return err
	}

	parts := []string{
		"", ctx.String("org"), ctx.String("project"), ctx.String("environment"),
		ctx.String("service"), identity, ctx.String("instance"),
	}

	fmt.Println("parts:")
	fmt.Println(parts)

	return nil
}

func listCmd(ctx *cli.Context) error {

	fmt.Println("")
	fmt.Println("")
	fmt.Println("*****************************************************************************")
	fmt.Println("              DEBUG OUTPUT   ")
	fmt.Println("")

	verbose := false
	if ctx.Bool("verbose"){
		verbose = true
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		//return errs.NewExitError("Failed to retrieve objects")
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Check to make sure these are valid
	parts := []string{"", ctx.String("org"), ctx.String("project"), "" }

	path := strings.Join(parts, "/")

	// Begin path tree
	// Create node for org
	treeHead := TreeNode{
		path: path,
		value: path,
		doDisplay: true,
		parent: nil,
	}

	fmt.Println("")
	fmt.Println("Listing secrets in project path:")
	fmt.Println(path)

	// Parse the partial pathexp to be used with matching
	pexp, err := pathexp.ParsePartial(path)
	if err != nil {
		return err
	}

	var projectTree registry.ProjectTreeSegment
	tree, err := projectTreeForOrg(c, client, pexp)
	if err != nil {
		return err
	}
	projectTree = *tree

	fmt.Println("")
	fmt.Println("Searching for environments...")
	for _, e := range projectTree.Envs {
		envPath := path + e.Body.Name + "/"
		fmt.Println("env: " + envPath)

		node := TreeNode{
			path: envPath,
			value: e.Body.Name,
			doDisplay: false,
			parent: &treeHead,
		}

		treeHead.children = append(treeHead.children, &node)

	}

	fmt.Println("")
	fmt.Println("Searching for services:")

	for _, e := range treeHead.children {

		for _, s := range projectTree.Services {
			sPath := e.path + s.Body.Name + "/"
			fmt.Println("service: " + sPath)

			node := TreeNode{
				path: sPath,
				value: s.Body.Name,
				doDisplay: false,
				parent: e,
			}

			secretPath := sPath + "*/*"
			creds, err := client.Credentials.Search(c, secretPath)
			if err != nil {
				return err
			}

			secrets := []apitypes.CredentialEnvelope{}
			for _, c := range creds{
				fmt.Println("spacingsecret: " + (*c.Body).GetName())

				secrets = append(secrets, c)
			}

			if len(secrets) != 0 {
				node.secrets = secrets
				node.doDisplay = true

				node.parent.doDisplay = true
				node.parent.parent.doDisplay = true
			}

			e.children = append(e.children, &node)
		}
	}

	// Print tree
	fmt.Println("")
	fmt.Println("")
	fmt.Println("*****************************************************************************")
	fmt.Println("")
	fmt.Println("")
	fmt.Println(treeHead.value)

	envTab := "  "
	serTab := "    "
	secTab := "      "
	for _, e := range treeHead.children{
		if e.doDisplay == false{
			continue
		}
		fmt.Println(envTab + e.value + "/")

		for _, s := range e.children{
			if s.doDisplay == false{
				continue
			}
			fmt.Println(serTab + s.value + "/")
			for _, sec := range s.secrets{
				if verbose{
					fmt.Println(secTab + (*sec.Body).GetPathExp().String() + "/" + (*sec.Body).GetName())
				} else {
					fmt.Println(secTab + (*sec.Body).GetName())
				}
			}
		}
	}

	return nil
}
