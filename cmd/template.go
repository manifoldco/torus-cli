package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/errs"
)

func init() {
	template := cli.Command{
		Name:      "template",
		Usage:     "Generate a config file from a template using secrets stored inside torus",
		ArgsUsage: "<template-file> [output-file]",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),
			userFlag("Use this user.", false),
			machineFlag("Use this machine.", false),
			stdInstanceFlag,
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, checkRequiredFlags, templateCmd,
		),
	}

	Cmds = append(Cmds, template)
}

func templateCmd(ctx *cli.Context) error {

	outputPath := ""
	args := ctx.Args()

	if len(args) == 0 {
		return errs.NewUsageExitError("A template-file must be provided", ctx)
	} else if len(args) > 2 {
		return errs.NewUsageExitError("Too many arguments provided", ctx)
	}

	templatePath, err := filepath.Abs(args[0])
	if err != nil {
		return errs.NewErrorExitError("Invalid template file path", err)
	}
	if len(args) == 2 {
		outputPath, err = filepath.Abs(args[1])
		if err != nil {
			return errs.NewErrorExitError("Invalid output file path", err)
		}
	}

	template, err := loadTemplateFile(templatePath)
	if err != nil {
		return err
	}

	secrets, _, err := getSecrets(ctx)
	if err != nil {
		return err
	}

	output, err := deriveTemplateOutput(template, secrets)
	if err != nil {
		return err
	}

	if outputPath != "" {
		return writeOutputFile(outputPath, output)
	}

	fmt.Printf("%s", output)
	return nil
}

func loadTemplateFile(templatePath string) (string, error) {
	buf, err := ioutil.ReadFile(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Could not open template file: %s", templatePath)
		}

		return "", err
	}

	return string(buf), nil
}

var templateFns = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()
	},
}

// We expose the variables in two ways to the template:
//
//  1) Through a _config variable which has a Name and Value property for iterating
//  2) Directly through the {{ slug }} and {{ Upper(slug) }} for straight referencing
func deriveTemplateOutput(body string, secrets []apitypes.CredentialEnvelope) (string, error) {
	variables := make(map[string]interface{})
	configItems := make([]map[string]string, len(secrets))

	for i, secret := range secrets {
		value := (*secret.Body).GetValue().String()
		name := (*secret.Body).GetName()
		configItems[i] = map[string]string{
			"name":  name,
			"value": value,
		}

		variables[name] = value
		variables[strings.ToUpper(name)] = value
	}

	// Export the config inside the template under `_config` without colliding
	// with the variable namespace
	variables["_config"] = configItems

	t := template.New("template").Funcs(templateFns).Option("missingkey=error")
	t, err := t.Parse(body)
	if err != nil {
		return "", errs.NewErrorExitError("Could not parse template", err)
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, variables)
	if err != nil {
		return "", errs.NewErrorExitError("Could not generate template", err)
	}

	return buf.String(), nil
}

func writeOutputFile(outputPath, output string) error {
	err := ioutil.WriteFile(outputPath, []byte(output), 0600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("File already exists at output path: %s", outputPath)
		}

		return err
	}

	return nil
}
