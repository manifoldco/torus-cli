package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/pathexp"
)

var setUnsetFlags = []cli.Flag{
	stdOrgFlag,
	stdProjectFlag,
	newSlicePlaceholder("environment, e", "ENV", "Use this environment.",
		"", "TORUS_ENVIRONMENT", true),
	newSlicePlaceholder("service, s", "SERVICE", "Use this service.",
		"default", "TORUS_SERVICE", true),
	newSlicePlaceholder("user, u", "USER", "Use this user.", "*", "TORUS_USER", false),
	newSlicePlaceholder("machine, m", "MACHINE", "Use this machine.", "*", "TORUS_MACHINE", false),
	newSlicePlaceholder("instance, i", "INSTANCE", "Use this instance.",
		"*", "TORUS_INSTANCE", true),
}

func init() {
	set := cli.Command{
		Name:      "set",
		Usage:     "Set a secret for a service and environment",
		ArgsUsage: "<name|path> <value> or <name|path>=<value>",
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
	key, value, err := parseSetArgs(args)

	if err != nil {
		return errs.NewUsageExitError(err.Error(), ctx)
	}

	path, cname, err := determinePath(ctx, key)
	if err != nil {
		return err
	}

	name := key
	if cname != nil {
		name = *cname
	}

	makers := valueMakers{}
	makers[name] = func() *apitypes.CredentialValue {
		return apitypes.NewStringCredentialValue(value)
	}

	s, p := spinner(fmt.Sprintf("Attempting to set credential %s", name))
	s.Start()
	_, err = setCredentials(ctx, path, makers, p)
	s.Stop()
	if err != nil {
		return errs.NewErrorExitError("Could not set credential.", err)
	}
	fmt.Printf("\nCredential %s has been set at %s/%s\n", name, path, name)

	hints.Display(hints.View, hints.Run, hints.Unset, hints.Import, hints.Export)
	return nil
}

// parseSetArgs returns a key and value from a list of arguments. If there is
// only one argument, try to parse using env var syntax: `KEY=VALUE`.
func parseSetArgs(args []string) (key string, value string, err error) {
	if len(args) == 1 {
		args = strings.SplitN(args[0], "=", 2)
	}

	if len(args) < 2 {
		return "", "", errors.New("A secret name and value must be supplied")
	} else if len(args) > 2 {
		return "", "", errors.New("Too many arguments were provided")
	}

	key = args[0]
	value = args[1]

	if key == "" || value == "" {
		return "", "", errors.New("A secret must have a name and value")
	}

	return key, value, nil
}

// determinePath returns a PathExp and a possible credential name if a full
// path was provided.
func determinePath(ctx *cli.Context, path string) (*pathexp.PathExp, *string, error) {
	// First try and use the cli args as a full path. it should override any
	// options.
	idx := strings.LastIndex(path, "/")
	name := path[idx+1:]

	// It looks like the user gave a path expression. use that instead of flags.
	var pe *pathexp.PathExp
	var err error
	if idx != -1 {
		path := path[:idx]
		pe, err = pathexp.ParsePartial(path)
		if err != nil {
			return nil, nil, errs.NewExitError(err.Error())
		}
		if name == "*" {
			return nil, nil, errs.NewExitError("Secret name cannot be wildcard")
		}
	} else {
		pe, err = determinePathFromFlags(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	return pe, &name, nil
}

func determinePathFromFlags(ctx *cli.Context) (*pathexp.PathExp, error) {
	// Falling back to flags. do the expensive population of the user flag now,
	// and see if any required flags (all of them) are missing.
	err := chain(setUserEnv, checkRequiredFlags)(ctx)
	if err != nil {
		return nil, err
	}

	identity, err := deriveIdentitySlice(ctx)
	if err != nil {
		return nil, err
	}

	return pathexp.New(
		ctx.String("org"),
		ctx.String("project"),
		ctx.StringSlice("environment"),
		ctx.StringSlice("service"),
		identity,
		ctx.StringSlice("instance"),
	)
}

type valueMaker func() *apitypes.CredentialValue
type valueMakers map[string]valueMaker

func setCredentials(ctx *cli.Context, pe *pathexp.PathExp, makers valueMakers, p api.ProgressFunc) ([]apitypes.CredentialEnvelope, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, err := client.Orgs.GetByName(c, pe.Org.String())
	if org == nil || err != nil {
		return nil, errs.NewExitError("Org not found")
	}

	pName := pe.Project.String()
	projects, err := listProjects(&c, client, org.ID, &pName)
	if len(projects) != 1 || err != nil {
		return nil, errs.NewExitError("Project not found")
	}
	project := projects[0]

	creds := []*apitypes.CredentialEnvelope{}
	for name, maker := range makers {
		value := maker()
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

		creds = append(creds, &apitypes.CredentialEnvelope{
			Version: 2,
			Body:    &cred,
		})
	}

	return client.Credentials.Create(c, creds, p)
}
