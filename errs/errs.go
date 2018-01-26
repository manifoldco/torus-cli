package errs

import (
	"regexp"

	"github.com/urfave/cli"
)

// Word without punctuation or space
var wordRegex = regexp.MustCompile(`\w`)

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return "Usage:\n" + spacer + ctx.App.HelpName + " " + ctx.Command.Name + " [command options] " + ctx.Command.ArgsUsage
}

// NewUsageExitError creates an ExitError with appended usage text
func NewUsageExitError(message string, ctx *cli.Context) error {
	if wordRegex.MatchString(message[len(message)-1:]) {
		message += "."
	}
	return cli.NewExitError(message+"\n"+usageString(ctx), -1)
}

// NewErrorExitError creates an ExitError with an appended error message
func NewErrorExitError(message string, err error) error {
	if wordRegex.MatchString(message[len(message)-1:]) {
		message += "."
	}
	return cli.NewExitError(message+"\n"+err.Error(), -1)
}

// NewExitError creates an ExitError with -1
func NewExitError(message string) error {
	if wordRegex.MatchString(message[len(message)-1:]) {
		message += "."
	}
	return cli.NewExitError(message, -1)
}

// MultiError takes a list of possible errors, filters out
// nil values, and returns a new cli.MultiError
func MultiError(errors ...error) cli.MultiError {
	nonNilErrs := []error{}
	for _, e := range errors {
		if e != nil {
			nonNilErrs = append(nonNilErrs, e)
		}
	}

	return cli.NewMultiError(nonNilErrs...)
}
