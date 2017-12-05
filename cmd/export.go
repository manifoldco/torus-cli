package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/errs"
)

var formatValues = []string{"env", "bash", "powershell", "fish", "cmd", "json", "tfvars"}
var formatDescription = "Format of exported secrets (" + strings.Join(formatValues, ", ") + ")"

type mod int

const (
	_ mod = iota
	quotes
	uppercase
)

func init() {
	export := cli.Command{
		Name:      "export",
		Usage:     "Export secrets for a specific environment and service inside a project",
		ArgsUsage: "[path to file] or use stdout redirection (e.g. `torus export > config.env`)",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),
			userFlag("Use this user.", false),
			machineFlag("Use this machine.", false),
			stdInstanceFlag,
			formatFlag(formatValues[0], formatDescription),
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, checkRequiredFlags, exportCmd,
		),
	}

	Cmds = append(Cmds, export)
}

func exportCmd(ctx *cli.Context) error {
	args := ctx.Args()
	filepath := ""
	if len(args) > 0 {
		if len(args) != 1 {
			return errs.NewUsageExitError("Only one argument can be supplied.", ctx)
		}

		filepath = args[0]
	}

	secrets, _, err := getSecrets(ctx)
	if err != nil {
		return err
	}

	format := ctx.String("format")
	if !validFormat(format) {
		return errs.NewUsageExitError(fmt.Sprintf("Invalid format provided: %s", format), ctx)
	}

	var w io.Writer = os.Stdout
	if filepath != "" {
		fd, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return errs.NewErrorExitError("Could not write to given filepath", err)
		}
		defer fd.Close()

		w = fd
	}

	switch format {
	case "env":
		err = writeFormat(w, secrets, "%s=%s\n", uppercase|quotes)
	case "bash":
		err = writeFormat(w, secrets, "export %s=%s\n", uppercase|quotes)
	case "powershell":
		err = writeFormat(w, secrets, "$Env:%s = \"%s\"\n", uppercase)
	case "cmd":
		err = writeFormat(w, secrets, "set %s=%s\n", quotes)
	case "fish":
		err = writeFormat(w, secrets, "set -x %s %s;\n", quotes)
	case "tfvars":
		err = writeFormat(w, secrets, "%s = %s\n", quotes)
	case "json":
		err = writeJSONFormat(w, secrets)
	default:
		return errs.NewUsageExitError(fmt.Sprintf("Could not find format for %s", format), ctx)
	}

	if err != nil {
		return errs.NewErrorExitError(fmt.Sprintf("Could not write format for %s", format), err)
	}

	return nil
}

func writeFormat(w io.Writer, secrets []apitypes.CredentialEnvelope, format string, modifier mod) error {
	tw := tabwriter.NewWriter(w, 2, 0, 2, ' ', 0)

	for _, secret := range secrets {
		name := (*secret.Body).GetName()
		value := (*secret.Body).GetValue().String()

		if modifier&quotes == quotes {
			value = fmt.Sprintf("%q", value)
		}

		if (modifier & uppercase) == uppercase {
			name = strings.ToUpper(name)
		}

		fmt.Fprintf(tw, format, name, value)
	}

	return tw.Flush()
}

func validFormat(format string) bool {
	for _, v := range formatValues {
		if v == format {
			return true
		}
	}

	return false
}
