package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/shlex"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/urfave/cli"
)

type secretPair struct {
	key   string
	value string
}

func init() {
	c := cli.Command{
		Name:      "import",
		Usage:     "Import multiple secrets from an env file",
		ArgsUsage: "[path to file] or use stdin redirection (e.g. `torus import < secrets.env`)",
		Category:  "SECRETS",
		Flags:     setUnsetFlags,
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setSliceDefaults, importCmd,
		),
	}

	Cmds = append(Cmds, c)
}

func importCmd(ctx *cli.Context) error {
	args := ctx.Args()
	secrets, err := importSecretFile(args)

	if err != nil {
		return errs.NewUsageExitError(err.Error(), ctx)
	}

	path, err := determinePathFromFlags(ctx)
	if err != nil {
		return err
	}

	makers := valueMakers{}
	for _, secret := range secrets {
		makers[secret.key] = func(value string) valueMaker {
			return func() *apitypes.CredentialValue {
				return apitypes.NewStringCredentialValue(value)
			}
		}(secret.value)
	}

	creds, err := setCredentials(ctx, path, makers)
	if err != nil {
		return errs.NewErrorExitError("Could not set credentials.", err)
	}

	fmt.Println()
	for _, cred := range creds {
		name := (*cred.Body).GetName()
		pe := (*cred.Body).GetPathExp()
		fmt.Printf("Credential %s has been set at %s/%s\n", name, pe, name)
	}

	hints.Display(hints.View, hints.Set)
	return nil
}

// importSecretFile returns a list of secret pairs either reading a file
// provided or from the standard input. It returns an error if there is a
// problem parsing secrets or stdin fails to read.
func importSecretFile(args []string) ([]secretPair, error) {
	switch len(args) {
	case 0:
		return readStdin()
	case 1:
		return readFile(args[0])
	default:
		return nil, errors.New("Too many arguments were provided")
	}
}

func readFile(filename string) ([]secretPair, error) {
	flags := os.O_RDONLY
	f, err := os.OpenFile(filename, flags, 0644)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading %s file %s", filename, err)
	}

	defer f.Close()

	return scanSecrets(f)
}

func readStdin() ([]secretPair, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("Could not from stdin. %s", err)
	}

	if (stat.Mode() & os.ModeNamedPipe) != 0 {
		return nil, errors.New("Could not read from piped input")
	}

	return scanSecrets(os.Stdin)
}

// scanSecrets reads secret pairs using an UNIX shell-like syntax parser. Empty
// lines and comments are ignored.
func scanSecrets(r io.Reader) ([]secretPair, error) {
	var pairs []secretPair

	lexer := shlex.NewLexer(r)

	for {
		word, err := lexer.Next()
		if err != nil {
			if err == io.EOF {
				return pairs, nil
			}
			return nil, fmt.Errorf("Error reading input file. %s", err)
		}

		tokens := strings.SplitN(word, "=", 2)
		if len(tokens) < 2 {
			return nil, fmt.Errorf("Error parsing secret %q", word)
		}

		key := tokens[0]
		value := tokens[1]

		if key == "" || value == "" {
			return nil, fmt.Errorf("Error parsing secret %q", word)
		}

		pairs = append(pairs, secretPair{key: key, value: value})
	}
}
