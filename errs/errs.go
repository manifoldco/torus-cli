package errs

import (
	"regexp"
	"strings"

	"github.com/urfave/cli"
)

// ErrAbort represents a situation where a user has been asked to perform an
// action but they've decided to abort.
var ErrAbort = NewExitError("Aborted.")

// ErrTerminalRequired represents a situation where a connected terminal is
// required to perform the action.
var ErrTerminalRequired = NewExitError("This action must be performed in an attached terminal or all arguments must be passed.")

// ToError converts the given error into a cli.ExitError
func ToError(err error) error {
	switch e := err.(type) {
	case *cli.ExitError:
		return e
	case nil:
		return nil
	default:
		return cli.NewExitError(err.Error(), -1)
	}
}

// NewNotFound returns an error representing a NotFound error
func NewNotFound(name string) error {
	return NewExitError("Could not find " + strings.ToLower(name))
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

// MultiError loops over all given errors unpacks them and then creates a new
// MultiError for non-nil entries.
func MultiError(errors ...error) cli.MultiError {
	nonNilErrs := []error{}
	for _, e := range errors {
		if e != nil {
			nonNilErrs = append(nonNilErrs, e)
		}
	}

	return cli.NewMultiError(nonNilErrs...)
}

func usageString(ctx *cli.Context) string {
	spacer := "    "
	return "Usage:\n" + spacer + ctx.App.HelpName + " " + ctx.Command.Name + " [command options] " + ctx.Command.ArgsUsage
}

// Word without punctuation or space
var wordRegex = regexp.MustCompile(`\w`)
