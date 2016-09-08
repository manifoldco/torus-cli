package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/pathexp"
)

func init() {
	set := cli.Command{
		Name:      "set",
		Usage:     "Set a secret for a service and environment",
		ArgsUsage: "<name|path> <value>",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			StdOrgFlag,
			StdProjectFlag,
			newSlicePlaceholder("environment, e", "ENV", "Use this environment.",
				nil, "AG_ENVIRONMENT", true),
			newSlicePlaceholder("service, s", "SERVICE", "Use this service.",
				[]string{"default"}, "AG_SERVICE", true),
			newSlicePlaceholder("user, u", "USER", "Use this user (identity).",
				[]string{"*"}, "AG_USER", true),
			newSlicePlaceholder("instance, i", "INSTANCE", "Use this instance.",
				[]string{"1"}, "AG_INSTANCE", true),
		},
		Action: Chain(
			EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults, setCmd,
		),
	}

	Cmds = append(Cmds, set)
}

func setCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		msg := "name and value are required.\n"
		if len(args) > 2 {
			msg = "Too many arguments provided.\n"
		}
		msg += usageString(ctx)
		return cli.NewExitError(msg, -1)
	}

	// First try and use the cli args as a full path. it should override any
	// options.
	idx := strings.LastIndex(args[0], "/")
	name := args[0][idx+1:]

	var pe *pathexp.PathExp

	// It looks like the user gave a path expression. use that instead of flags.
	if idx != -1 {
		var err error
		path := args[0][:idx]
		pe, err = pathexp.Parse(path)
		if err != nil {
			return cli.NewExitError("Error reading path expression: "+err.Error(), -1)
		}
	} else {
		// Falling back to flags. do the expensive population of the user flag now,
		// and see if any required flags (all of them) are missing.
		err := Chain(SetUserEnv, checkRequiredFlags)(ctx)
		if err != nil {
			return err
		}

		pe, err = pathexp.New(ctx.String("org"), ctx.String("project"),
			ctx.StringSlice("environment"), ctx.StringSlice("service"),
			ctx.StringSlice("user"), ctx.StringSlice("instance"),
		)
		if err != nil {
			return err
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := client.Orgs.GetByName(c, pe.Org())
	if org == nil || err != nil {
		return cli.NewExitError("Org not found", -1)
	}

	pName := pe.Project()
	projects, err := client.Projects.List(c, org.ID, &pName)
	if len(projects) != 1 || err != nil {
		return cli.NewExitError("Project not found", -1)
	}
	project := projects[0]

	var v *apitypes.CredentialValue
	if i, err := strconv.Atoi(args[1]); err == nil {
		v = apitypes.NewIntCredentialValue(i)
	} else if f, err := strconv.ParseFloat(args[1], 64); err == nil {
		v = apitypes.NewFloatCredentialValue(f)
	} else {
		v = apitypes.NewStringCredentialValue(args[1])
	}

	cred := apitypes.Credential{
		OrgID:     org.ID,
		ProjectID: project.ID,
		Name:      name,
		PathExp:   pe,
		Value:     v,
	}

	_, err = client.Credentials.Create(c, &cred, &progress)
	if err != nil {
		return cli.NewExitError("Could not set credential: "+err.Error(), -1)
	}

	fmt.Printf("\nCredential %s has been set at %s/%s\n", name, pe, name)

	return nil
}
