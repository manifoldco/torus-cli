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
	"github.com/arigatomachine/cli/errs"
	"github.com/arigatomachine/cli/pathexp"
)

var setUnsetFlags = []cli.Flag{
	stdOrgFlag,
	stdProjectFlag,
	newSlicePlaceholder("environment, e", "ENV", "Use this environment.",
		"", "TORUS_ENVIRONMENT", true),
	newSlicePlaceholder("service, s", "SERVICE", "Use this service.",
		"default", "TORUS_SERVICE", true),
	newSlicePlaceholder("user, u", "USER", "Use this user (identity).",
		"*", "TORUS_USER", true),
	newSlicePlaceholder("instance, i", "INSTANCE", "Use this instance.",
		"*", "TORUS_INSTANCE", true),
}

func init() {
	set := cli.Command{
		Name:      "set",
		Usage:     "Set a secret for a service and environment",
		ArgsUsage: "<name|path> <value>",
		Category:  "SECRETS",
		Flags:     setUnsetFlags,
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setSliceDefaults, setCmd,
		),
	}

	Cmds = append(Cmds, set)
}

func setCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		msg := "name and value are required."
		if len(args) > 2 {
			msg = "Too many arguments provided."
		}
		return errs.NewUsageExitError(msg, ctx)
	}

	cred, err := setCredential(ctx, args[0], func() *apitypes.CredentialValue {
		var v *apitypes.CredentialValue
		if i, err := strconv.Atoi(args[1]); err == nil {
			v = apitypes.NewIntCredentialValue(i)
		} else if f, err := strconv.ParseFloat(args[1], 64); err == nil {
			v = apitypes.NewFloatCredentialValue(f)
		} else {
			v = apitypes.NewStringCredentialValue(args[1])
		}

		return v
	})

	if err != nil {
		return errs.NewErrorExitError("Could not set credential.", err)
	}

	name := (*cred.Body).GetName()
	pe := (*cred.Body).GetPathExp()
	fmt.Printf("\nCredential %s has been set at %s/%s\n", name, pe, name)

	return nil
}

func setCredential(ctx *cli.Context, nameOrPath string, valueMaker func() *apitypes.CredentialValue) (*apitypes.CredentialEnvelope, error) {
	// First try and use the cli args as a full path. it should override any
	// options.
	idx := strings.LastIndex(nameOrPath, "/")
	name := nameOrPath[idx+1:]

	var pe *pathexp.PathExp

	// It looks like the user gave a path expression. use that instead of flags.
	if idx != -1 {
		var err error
		path := nameOrPath[:idx]
		pe, err = pathexp.Parse(path)
		if err != nil {
			return nil, errs.NewExitError(err.Error())
		}
	} else {
		// Falling back to flags. do the expensive population of the user flag now,
		// and see if any required flags (all of them) are missing.
		err := chain(setUserEnv, checkRequiredFlags)(ctx)
		if err != nil {
			return nil, err
		}

		pe, err = pathexp.New(ctx.String("org"), ctx.String("project"),
			ctx.StringSlice("environment"), ctx.StringSlice("service"),
			ctx.StringSlice("user"), ctx.StringSlice("instance"),
		)
		if err != nil {
			return nil, err
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := client.Orgs.GetByName(c, pe.Org())
	if org == nil || err != nil {
		return nil, errs.NewExitError("Org not found")
	}

	pName := pe.Project()
	projects, err := listProjects(&c, client, org.ID, &pName)
	if len(projects) != 1 || err != nil {
		return nil, errs.NewExitError("Project not found")
	}
	project := projects[0]
	value := valueMaker()

	state := "set"
	if value.IsUnset() {
		state = "unset"
		value = nil
	}

	var cred apitypes.Credential
	cBodyV2 := apitypes.CredentialV2{
		BaseCredential: apitypes.BaseCredential{
			OrgID:     org.ID,
			ProjectID: project.ID,
			Name:      strings.ToLower(name),
			PathExp:   pe,
			Value:     value,
		},
		State: state,
	}
	cred = &cBodyV2

	return client.Credentials.Create(c, &cred, &progress)
}
