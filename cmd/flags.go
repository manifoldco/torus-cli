package cmd

// Standard flag definitions shared across commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// Standard flags with generic descriptions
var (
	stdOrgFlag     = orgFlag("Use this organization.", true)
	stdProjectFlag = projectFlag("Use this project.", true)
	stdEnvFlag     = envFlag("Use this environment.", true)

	stdAutoAcceptFlag = cli.BoolFlag{
		Name:  "yes, y",
		Usage: "Automatically accept confirmation dialogues.",
	}
)

// formatFlag creates a new --format cli.Flag with custom usage string
func formatFlag(defaultValue, description string) cli.Flag {
	return newPlaceholder("format, f", "FORMAT", description,
		defaultValue, "TORUS_FORMAT", false)
}

// nameFlag creates a new --name cli.Flag with custom usage string
func nameFlag(description string) cli.Flag {
	return newPlaceholder("name, n", "NAME", description, "", "TORUS_NAME", false)
}

// descriptionFlag creates a new --description, -d cli.Flag with custom usage string
func descriptionFlag(description string) cli.Flag {
	return newPlaceholder("description, d", "DESCRIPTION", description, "", "TORUS_DESCRIPTION", false)
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

// envSliceFlag creates a new --environment cli.StringSliceFlag with custom usage string.
func envSliceFlag(usage string, required bool) cli.Flag {
	return newSlicePlaceholder("environment, e", "ENV", usage, "", "TORUS_ENVIRONMENT", required)
}

// serviceFlag creates a new --service cli.Flag with custom usage string.
func serviceFlag(usage, value string, required bool) cli.Flag {
	return newPlaceholder("service, s", "SERVICE", usage, value, "TORUS_SERVICE", required)
}

// serviceSliceFlag creates a new --service cli.StringSliceFlag with custom usage string.
func serviceSliceFlag(usage, value string, required bool) cli.Flag {
	return newSlicePlaceholder("service, s", "SERVICE", usage, value, "TORUS_SERVICE", required)
}

// machineFlag creates a new --machine cli.Flag with custom usage string.
func machineFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("machine, m", "MACHINE", usage, "", "TORUS_MACHINE", required)
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
