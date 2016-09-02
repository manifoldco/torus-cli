package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
	"gopkg.in/oleiade/reflections.v1"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/dirprefs"
	"github.com/arigatomachine/cli/prefs"
)

// Chain allows easy sequential calling of BeforeFuncs and AfterFuncs.
// Chain will exit on the first error seen.
// XXX Chain is only public while we need it for passthrough.go
func Chain(funcs ...func(*cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {

		for _, f := range funcs {
			err := f(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// EnsureDaemon ensures that the daemon is running, and is the correct version,
// before a command is exeucted.
// the daemon will be started/restarted once, to try and launch the latest
// version.
// XXX EnsureDaemon is only public while we need it for passthrough.go
func EnsureDaemon(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	proc, err := findDaemon(cfg)
	if err != nil {
		return err
	}

	spawned := false

	if proc == nil {
		err := spawnDaemon()
		if err != nil {
			return err
		}

		spawned = true
	}

	client := api.NewClient(cfg)

	var v *apitypes.Version
	increment := 5 * time.Millisecond
	for d := increment; d < 1*time.Second; d += increment {
		v, err = client.Version.Get(context.Background())
		if err == nil {
			break
		}
		time.Sleep(d)
	}

	if err != nil {
		return cli.NewExitError("Error communicating with the daemon: "+err.Error(), -1)
	}

	if v.Version == cfg.Version {
		return nil
	}

	if spawned {
		return cli.NewExitError("The daemon version is incorrect. Check for stale processes", -1)
	}

	fmt.Println("The daemon version is out of date and is being restarted.")
	fmt.Println("You will need to login again.")

	_, err = stopDaemon(proc)
	if err != nil {
		return err
	}

	return EnsureDaemon(ctx)
}

// EnsureSession ensures that the user is logged in with the daemon and has a
// valid session. If not, it will attempt to log the user in via environment
// variables. If they do not exist, of the login fails, it will abort the
// command.
// XXX EnsureSession is only public while we need it for passthrough.go
func EnsureSession(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	_, err = client.Session.Get(context.Background())

	hasSession := true
	if err != nil {
		if cerr, ok := err.(*apitypes.Error); ok {
			if cerr.Type == apitypes.UnauthorizedError {
				hasSession = false
			}
		}
		if hasSession {
			return cli.NewExitError("Error communicating with the daemon: "+err.Error(), -1)
		}
	}

	if hasSession {
		return nil
	}

	email, hasEmail := os.LookupEnv("AG_EMAIL")
	password, hasPassword := os.LookupEnv("AG_PASSWORD")

	if hasEmail && hasPassword {
		fmt.Println("Attempting to login with email: " + email)

		err := client.Session.Login(context.Background(), email, password)
		if err != nil {
			fmt.Println("Could not log in: " + err.Error())
		} else {
			return nil
		}
	}

	msg := "You must be logged in to run '" + ctx.Command.FullName() + "'.\n" +
		"Login using 'login' or create an account using 'signup'."
	return cli.NewExitError(msg, -1)
}

// LoadDirPrefs loads argument values from the .arigato.json file
// XXX LoadDirPrefs is only public while we need it for passthrough.go
func LoadDirPrefs(ctx *cli.Context) error {
	p, err := prefs.NewPreferences(true)
	if err != nil {
		return err
	}

	d, err := dirprefs.Load()
	if err != nil {
		return err
	}

	return reflectArgs(ctx, p, d, "json")
}

// LoadPrefDefaults loads default argument values from the .arigatorc
// preferences file defaults section, inserting them into any unset flag values
// XXX LoadPrefDefaults is only public while we need it for passthrough.go
func LoadPrefDefaults(ctx *cli.Context) error {
	p, err := prefs.NewPreferences(true)
	if err != nil {
		return err
	}

	return reflectArgs(ctx, p, p.Defaults, "ini")
}

func reflectArgs(ctx *cli.Context, p *prefs.Preferences, i interface{},
	tagName string) error {

	// The user has disabled reading arguments from prefs and .arigato.json
	if !p.Core.Context {
		return nil
	}

	// tagged field names match the argument names
	tags, err := reflections.Tags(i, tagName)
	if err != nil {
		return err
	}

	flags := make(map[string]bool)
	for _, flagName := range ctx.FlagNames() {
		// This value is already set via arguments or env vars. skip it.
		if ctx.String(flagName) != "" {
			continue
		}

		flags[flagName] = true
	}

	for fieldName, tag := range tags {
		name := strings.SplitN(tag, ",", 2)[0] // remove omitempty if its there
		if _, ok := flags[name]; ok {
			field, err := reflections.GetField(i, fieldName)
			if err != nil {
				return err
			}
			ctx.Set(name, field.(string))
		}
	}

	return nil
}

// SetUserEnv populates the env argument, if present and unset,
// with dev-USERNAME
// XXX SetUserEnv is only public while we need it for passthrough.go
func SetUserEnv(ctx *cli.Context) error {
	argName := "environment"
	// Check for env flag, just in case this middleware is misused
	hasEnvFlag := false
	for _, name := range ctx.FlagNames() {
		if name == argName {
			hasEnvFlag = true
			break
		}
	}
	if !hasEnvFlag {
		return nil
	}

	env := ctx.String(argName)
	if env != "" {
		return nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	u, err := client.Users.Self(context.Background())
	if err != nil {
		return err
	}

	ctx.Set(argName, "dev-"+u.Body.Username)
	return nil
}

// CheckRequiredFlags ensures that any required flags have been set either on
// the command line, or through envvars/prefs files.
func checkRequiredFlags(ctx *cli.Context) error {
	missing := []string{}
	for _, f := range ctx.Command.Flags {
		if psf, ok := f.(placeHolderStringFlag); ok {
			name := strings.SplitN(psf.GetName(), ",", 2)[0]
			if psf.Required && ctx.String(name) == "" {
				prefix := "-"
				if len(name) > 1 {
					prefix = "--"
				}
				missing = append(missing, prefix+name)
			}
		}
	}

	if len(missing) > 0 {
		msg := "Missing flags: " + strings.Join(missing, ", ") + "\n"
		msg += usageString(ctx)
		return cli.NewExitError(msg, -1)
	}

	return nil
}
