package cmd

// Standard flag definitions shared across commands
// XXX: These are public for use in passthrough right now.

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// Standard flags for path expression parts.
var (
	StdOrgFlag      = OrgFlag("Use this organization.")
	StdProjectFlag  = ProjectFlag("Use this project.")
	StdEnvFlag      = EnvFlag("Use this environment.")
	StdServiceFlag  = ServiceFlag("Use this service.")
	StdUserFlag     = UserFlag("Use this user.")
	StdInstanceFlag = InstanceFlag("Use this instance.")
)

// OrgFlag creates a new --org cli.Flag with custom usage string.
func OrgFlag(usage string) cli.Flag {
	return newPlaceholder("org, o", "ORG", usage, "", "AG_ORG")
}

// ProjectFlag creates a new --project cli.Flag with custom usage string.
func ProjectFlag(usage string) cli.Flag {
	return newPlaceholder("project, p", "PROJECT", usage, "", "AG_PROJECT")
}

// EnvFlag creates a new --environment cli.Flag with custom usage string.
func EnvFlag(usage string) cli.Flag {
	return newPlaceholder("environment, e", "ENV", usage, "", "AG_ENVIRONMENT")
}

// ServiceFlag creates a new --service cli.Flag with custom usage string.
func ServiceFlag(usage string) cli.Flag {
	return newPlaceholder("service, s", "SERVICE", usage, "", "AG_SERVICE")
}

// UserFlag creates a new --user cli.Flag with custom usage string.
func UserFlag(usage string) cli.Flag {
	return newPlaceholder("user, u", "USER", usage, "", "AG_USER")
}

// InstanceFlag creates a new --instance cli.Flag with custom usage string.
func InstanceFlag(usage string) cli.Flag {
	return newPlaceholder("instance, i", "INSTANCE", usage, "1", "AG_INSTANCE")
}

// placeHolderStringSliceFlag is a StringSliceFlag that has been extended to use a
// specific placedholder value in the usage, without parsing it out of the
// usage string.
type placeHolderStringSliceFlag struct {
	cli.StringSliceFlag
	Placeholder string
}

func (psf placeHolderStringSliceFlag) String() string {
	flags := prefixedNames(psf.Name, psf.Placeholder)
	def := ""
	if psf.Value != nil && len([]string(*psf.Value)) > 0 {
		def = fmt.Sprintf(" (default: %s)", psf.Value)
	}
	return fmt.Sprintf("%s\t%s%s", flags, psf.Usage, def)
}

func newSlicePlaceholder(name, placeholder, usage string, value cli.StringSlice) placeHolderStringSliceFlag {
	return placeHolderStringSliceFlag{
		StringSliceFlag: cli.StringSliceFlag{
			Name:  name,
			Usage: usage,
			Value: &value,
		},
		Placeholder: placeholder,
	}
}

// placeHolderStringFlag is a StringFlag that has been extended to use a
// specific placedholder value in the usage, without parsing it out of the
// usage string.
type placeHolderStringFlag struct {
	cli.StringFlag
	Placeholder string
}

func (psf placeHolderStringFlag) String() string {
	flags := prefixedNames(psf.Name, psf.Placeholder)
	def := ""
	if psf.Value != "" {
		def = fmt.Sprintf(" (default: %s)", psf.Value)
	}
	return fmt.Sprintf("%s\t%s%s", flags, psf.Usage, def)
}

func newPlaceholder(name, placeholder, usage, value, envvar string) placeHolderStringFlag {
	return placeHolderStringFlag{
		StringFlag: cli.StringFlag{
			Name:   name,
			Usage:  usage,
			Value:  value,
			EnvVar: envvar,
		},
		Placeholder: placeholder,
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
