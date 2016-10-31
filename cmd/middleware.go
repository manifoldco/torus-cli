package cmd

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/urfave/cli"
	"gopkg.in/oleiade/reflections.v1"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/dirprefs"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
)

// chain allows easy sequential calling of BeforeFuncs and AfterFuncs.
// chain will exit on the first error seen.
func chain(funcs ...func(*cli.Context) error) func(*cli.Context) error {
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

// ensureDaemon ensures that the daemon is running, and is the correct version,
// before a command is exeucted.
// the daemon will be started/restarted once, to try and launch the latest
// version.
func ensureDaemon(ctx *cli.Context) error {
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
		return errs.NewErrorExitError("Could not communicate with daemon.", err)
	}

	if v.Version == cfg.Version {
		return nil
	}

	if spawned {
		return errs.NewExitError("The daemon version is incorrect. Check for stale processes.")
	}

	fmt.Println("The daemon version is out of date and is being restarted.")
	fmt.Println("You will need to login again.")

	_, err = stopDaemon(proc)
	if err != nil {
		return err
	}

	return ensureDaemon(ctx)
}

// ensureSession ensures that the user is logged in with the daemon and has a
// valid session. If not, it will attempt to log the user in via environment
// variables. If they do not exist, of the login fails, it will abort the
// command.
func ensureSession(ctx *cli.Context) error {
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
			return errs.NewErrorExitError("Could not communicate with daemon.", err)
		}
	}

	if hasSession {
		return nil
	}

	email, hasEmail := os.LookupEnv("TORUS_EMAIL")
	password, hasPassword := os.LookupEnv("TORUS_PASSWORD")

	if hasEmail && hasPassword {
		fmt.Println("Attempting to login with email: " + email)

		err := client.Session.Login(context.Background(), email, password)
		if err != nil {
			fmt.Println("Could not log in.\n" + err.Error())
		} else {
			return nil
		}
	}

	msg := "You must be logged in to run '" + ctx.Command.FullName() + "'.\n" +
		"Login using 'login' or create an account using 'signup'."
	return errs.NewExitError(msg)
}

// loadDirPrefs loads argument values from the .torus.json file
func loadDirPrefs(ctx *cli.Context) error {
	p, err := prefs.NewPreferences(true)
	if err != nil {
		return err
	}

	d, err := dirprefs.Load(true)
	if err != nil {
		return err
	}

	return reflectArgs(ctx, p, d, "json")
}

// loadPrefDefaults loads default argument values from the .torusrc
// preferences file defaults section, inserting them into any unset flag values
func loadPrefDefaults(ctx *cli.Context) error {
	p, err := prefs.NewPreferences(true)
	if err != nil {
		return err
	}

	return reflectArgs(ctx, p, p.Defaults, "ini")
}

func reflectArgs(ctx *cli.Context, p *prefs.Preferences, i interface{},
	tagName string) error {

	// The user has disabled reading arguments from prefs and .torus.json
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
		if isSet(ctx, flagName) {
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

			if f, ok := field.(string); ok && f != "" {
				ctx.Set(name, field.(string))
			}
		}
	}

	return nil
}

// setUserEnv populates the env argument, if present and unset,
// with dev-USERNAME
func setUserEnv(ctx *cli.Context) error {
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

	if isSet(ctx, argName) {
		return nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	session, err := client.Session.Who(context.Background())
	if err != nil {
		return err
	}

	if session.Type() == apitypes.UserSession {
		ctx.Set(argName, "dev-"+session.Username())
	}

	return nil
}

// setSliceDefaults populates any string slice flags with the default value
// if nothing else is set. This is different from the default urfave default
// Value, which will always be included in the string slice options.
func setSliceDefaults(ctx *cli.Context) error {
	for _, f := range ctx.Command.Flags {
		if psf, ok := f.(placeHolderStringSliceFlag); ok {
			name := strings.SplitN(psf.GetName(), ",", 2)[0]
			if psf.Default != "" && len(ctx.StringSlice(name)) == 0 {
				ctx.Set(name, psf.Default)
			}
		}
	}

	return nil
}

func isSet(ctx *cli.Context, name string) bool {
	value := ctx.Generic(name)
	if value != nil {
		v := reflect.Indirect(reflect.ValueOf(value))
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			return v.Len() != 0
		}

		return true
	}

	return false
}

// CheckRequiredFlags ensures that any required flags have been set either on
// the command line, or through envvars/prefs files.
func checkRequiredFlags(ctx *cli.Context) error {
	missing := []string{}
	for _, f := range ctx.Command.Flags {
		var name string
		flagMissing := false
		switch pf := f.(type) {
		case placeHolderStringFlag:
			name = strings.SplitN(pf.GetName(), ",", 2)[0]
			if pf.Required && ctx.String(name) == "" {
				flagMissing = true
			}
		case placeHolderStringSliceFlag:
			name = strings.SplitN(pf.GetName(), ",", 2)[0]
			if pf.Required && len(ctx.StringSlice(name)) == 0 {
				flagMissing = true
			}

		}
		if flagMissing {
			prefix := "-"
			if len(name) > 1 {
				prefix = "--"
			}
			missing = append(missing, prefix+name)
		}

	}

	if len(missing) > 0 {
		msg := "Missing flags: " + strings.Join(missing, ", ")
		return errs.NewUsageExitError(msg, ctx)
	}

	return nil
}
