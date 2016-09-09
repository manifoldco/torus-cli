package cmd

// Standard flag definitions shared across commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// Standard flags for path expression parts.
var (
	stdOrgFlag      = orgFlag("Use this organization.", true)
	stdProjectFlag  = projectFlag("Use this project.", true)
	stdEnvFlag      = envFlag("Use this environment.", true)
	stdServiceFlag  = serviceFlag("Use this service.", "", true)
	stdUserFlag     = userFlag("Use this user.", true)
	stdInstanceFlag = instanceFlag("Use this instance.", true)
)

// orgFlag creates a new --org cli.Flag with custom usage string.
func orgFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("org, o", "ORG", usage, "", "AG_ORG", required)
}

// projectFlag creates a new --project cli.Flag with custom usage string.
func projectFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("project, p", "PROJECT", usage, "", "AG_PROJECT", required)
}

// envFlag creates a new --environment cli.Flag with custom usage string.
func envFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("environment, e", "ENV", usage, "", "AG_ENVIRONMENT", required)
}

// serviceFlag creates a new --service cli.Flag with custom usage string.
func serviceFlag(usage, value string, required bool) cli.Flag {
	return newPlaceholder("service, s", "SERVICE", usage, value, "AG_SERVICE", required)
}

// userFlag creates a new --user cli.Flag with custom usage string.
func userFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("user, u", "USER", usage, "", "AG_USER", required)
}

// instanceFlag creates a new --instance cli.Flag with custom usage string.
func instanceFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("instance, i", "INSTANCE", usage, "1", "AG_INSTANCE", required)
}

// placeHolderStringSliceFlag is a StringSliceFlag that has been extended to use a
// specific placedholder value in the usage, without parsing it out of the
// usage string.
type placeHolderStringSliceFlag struct {
	cli.StringSliceFlag
	Placeholder string
	Required    bool
}

func (psf placeHolderStringSliceFlag) String() string {
	flags := prefixedNames(psf.Name, psf.Placeholder)
	def := ""
	if psf.Value != nil && len(*psf.Value) > 0 {
		def = fmt.Sprintf(" (default: %s)", strings.Join(*psf.Value, ", "))
	}

	multi := " Can be specified multiple times."
	if psf.Usage[len(psf.Usage)-1] != '.' {
		multi = "." + multi
	}

	return fmt.Sprintf("%s\t%s%s%s", flags, psf.Usage, multi, def)
}

func newSlicePlaceholder(name, placeholder, usage string, value cli.StringSlice,
	envvar string, required bool) placeHolderStringSliceFlag {

	return placeHolderStringSliceFlag{
		StringSliceFlag: cli.StringSliceFlag{
			Name:   name,
			Usage:  usage,
			Value:  &value,
			EnvVar: envvar,
		},
		Placeholder: placeholder,
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
