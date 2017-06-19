package cmd

// Standard flag definitions shared across commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/errs"
)

// Standard flags with generic descriptions
var (
	stdOrgFlag      = orgFlag("Use this organization.", true)
	stdProjectFlag  = projectFlag("Use this project.", true)
	stdEnvFlag      = envFlag("Use this environment.", true)
	stdInstanceFlag = instanceFlag("Use this instance.", true)

	stdAutoAcceptFlag = cli.BoolFlag{
		Name:  "yes, y",
		Usage: "Automatically accept confirmation dialogues.",
	}

)

func formatFlag(defaultValue, description string) cli.Flag {
	return newPlaceholder("format, f", "FORMAT", description,
		defaultValue, "TORUS_FORMAT", false)
}

// orgFlag creates a new --org cli.Flag with custom usage string.
func orgFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("org, o", "ORG", usage, "", "TORUS_ORG", required)
}

// projectFlag creates a new --project cli.Flag with custom usage string.
func projectFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("project, p", "PROJECT", usage, "", "TORUS_PROJECT", required)
}

// envFlag creates a new --environment cli.Flag with custom usage string.
func envFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("environment, e", "ENV", usage, "", "TORUS_ENVIRONMENT", required)
}

// serviceFlag creates a new --service cli.Flag with custom usage string.
func serviceFlag(usage, value string, required bool) cli.Flag {
	return newPlaceholder("service, s", "SERVICE", usage, value, "TORUS_SERVICE", required)
}

// userFlag creates a new --user cli.Flag with custom usage string.
func userFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("user, u", "USER", usage, "", "TORUS_USER", required)
}

// machineFlag creates a new --machine cli.Flag with custom usage string.
func machineFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("machine, m", "MACHINE", usage, "", "TORUS_MACHINE", required)
}

// instanceFlag creates a new --instance cli.Flag with custom usage string.
func instanceFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("instance, i", "INSTANCE", usage, "1", "TORUS_INSTANCE", required)
}

// roleFlag creates a new --role cli.Flag with custom usage string.
func roleFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("role, r", "ROLE", usage, "", "TORUS_ROLE", required)
}

// destroyedFlag creates a new --destroyed cli.Flag with custom usage string.
func destroyedFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "destroyed, d",
		Usage: "Display only destroyed",
	}
}

// placeHolderStringSliceFlag is a StringSliceFlag that has been extended to use a
// specific placedholder value in the usage, without parsing it out of the
// usage string.
type placeHolderStringSliceFlag struct {
	cli.StringSliceFlag
	Placeholder string
	Default     string
	Required    bool
}

func (psf placeHolderStringSliceFlag) String() string {
	flags := prefixedNames(psf.Name, psf.Placeholder)
	def := ""
	if len(psf.Default) > 0 {
		def = fmt.Sprintf(" (default: %s)", psf.Default)
	}

	multi := " Can be specified multiple times."
	if psf.Usage[len(psf.Usage)-1] != '.' {
		multi = "." + multi
	}

	return fmt.Sprintf("%s\t%s%s%s", flags, psf.Usage, multi, def)
}

func newSlicePlaceholder(name, placeholder, usage string, value string,
	envvar string, required bool) placeHolderStringSliceFlag {

	return placeHolderStringSliceFlag{
		StringSliceFlag: cli.StringSliceFlag{
			Name:   name,
			Usage:  usage,
			EnvVar: envvar,
		},
		Placeholder: placeholder,
		Default:     value,
		Required:    required,
	}
}

// placeHolderStringFlag is a StringFlag that has been extended to use a
// specific placedholder value in the usage, without parsing it out of the
// usage string.
type placeHolderStringFlag struct {
	cli.StringFlag
	Placeholder string
	Required    bool
}

func (psf placeHolderStringFlag) String() string {
	flags := prefixedNames(psf.Name, psf.Placeholder)
	def := ""
	if psf.Value != "" {
		def = fmt.Sprintf(" (default: %s)", psf.Value)
	}
	return fmt.Sprintf("%s\t%s%s", flags, psf.Usage, def)
}

func newPlaceholder(name, placeholder, usage, value, envvar string,
	required bool) placeHolderStringFlag {

	return placeHolderStringFlag{
		StringFlag: cli.StringFlag{
			Name:   name,
			Usage:  usage,
			Value:  value,
			EnvVar: envvar,
		},
		Placeholder: placeholder,
		Required:    required,
	}
}

// prefixedNames and prefixFor are taken from urfave/cli
func prefixedNames(fullName, placeholder string) string {
	var prefixed string
	parts := strings.Split(fullName, ",")
	for i, name := range parts {
		name = strings.Trim(name, " ")
		prefixed += prefixFor(name) + name
		if placeholder != "" {
			prefixed += " " + placeholder
		}
		if i < len(parts)-1 {
			prefixed += ", "
		}
	}
	return prefixed
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}

func identityString(identityType, identity string) (string, error) {
	if identityType == "user" {
		return identity, nil
	}

	if strings.HasPrefix(identity, "machine-") {
		return "", errs.NewExitError(
			"Do not prepend 'machine-' when using --machine")
	}

	return "machine-" + identity, nil
}

func deriveIdentity(ctx *cli.Context, session *api.Session) (string, error) {
	if ctx.String("user") != "" && ctx.String("machine") != "" {
		return "", errs.NewExitError(
			"You can only supply --user or --machine, not both.")
	}

	identity := ""
	if ctx.String("user") != "" {
		identity = ctx.String("user")
	}

	var err error
	if ctx.String("machine") != "" {
		identity, err = identityString("machine", ctx.String("machine"))
		if err != nil {
			return "", err
		}
	}

	if identity == "" {
		identity, err = identityString(string(session.Type()), session.Username())
		if err != nil {
			return "", err
		}
	}

	return identity, nil
}

// Derives a slice of identity segments for us in building a PathExp object.
func deriveIdentitySlice(ctx *cli.Context) ([]string, error) {
	users := ctx.StringSlice("user")
	machines := ctx.StringSlice("machine")

	var identities []string
	for _, u := range users {
		identity, err := identityString("user", u)
		if err != nil {
			return identities, err
		}

		if identity == "*" {
			return []string{"*"}, nil
		}

		identities = append(identities, identity)
	}

	for _, m := range machines {
		identity, err := identityString("machine", m)
		if err != nil {
			return identities, err
		}

		if identity == "*" {
			return []string{"*"}, nil
		}

		identities = append(identities, identity)
	}

	return identities, nil
}
