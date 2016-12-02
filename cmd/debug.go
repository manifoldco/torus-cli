package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
)

func init() {
	version := cli.Command{
		Name:  "debug",
		Usage: "Display useful debug information for submission to support",
		Flags: []cli.Flag{
			// These flags are hidden so we can still parse out the values
			// from the prefs files and env vars, but we don't display
			// them to the users in help.
			// A user could still set the flag on the command line though :(
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "environment", EnvVar: "TORUS_ENVIRONMENT", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "service", Value: "default", EnvVar: "TORUS_SERVICE", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "org", EnvVar: "TORUS_ORG", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "project", EnvVar: "TORUS_PROJECT", Hidden: true},
				Required:   true,
			},
			placeHolderStringFlag{
				StringFlag: cli.StringFlag{Name: "instance", EnvVar: "TORUS_INSTANCE", Value: "1", Hidden: true},
				Required:   true,
			},
		},
		Action: chain(ensureDaemon, loadDirPrefs, loadPrefDefaults, setUserEnv, debugInfoCmd),
		Hidden: true,
	}
	Cmds = append(Cmds, version)
}

func debugInfoCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	timestamp := time.Now()

	// Cli and registry versions
	daemonVersion, registryVersion := retrieveVersions(c, client)

	// User information
	session, uErr := client.Session.Who(c)
	loggedIn := true
	if uErr != nil {
		if strings.Contains(uErr.Error(), "invalid login") {
			loggedIn = false
		} else {
			return errs.NewErrorExitError("Could not retrieve user", err)
		}
	}

	// Debug environment variable used
	debug := false
	if os.Getenv("TORUS_DEBUG") != "" {
		debug = true
	}

	// Which registry
	var registryURI string
	preferences, err := prefs.NewPreferences()
	if err != nil {
		registryURI = "Failed to load prefs"
	} else {
		registryURI = preferences.Core.RegistryURI
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", "Timestamp", timestamp.UTC().Format(time.UnixDate))
	fmt.Fprintf(w, "%s\t%v\n", "Debug", debug)

	if !loggedIn {
		fmt.Fprintf(w, "%s\t%v\n", "Logged In", loggedIn)
	} else {
		fmt.Fprintf(w, "%s\t%s\n", "Session Type", session.Type())

		if session.Type() == apitypes.MachineSession {
			fmt.Fprintf(w, "%s\t%s\n", "Machine ID", session.ID())
			fmt.Fprintf(w, "%s\t%s\n", "Machine Token ID", session.AuthID())
			fmt.Fprintf(w, "%s\t%s\n", "Machine Name", session.Username())
		} else {
			fmt.Fprintf(w, "%s\t%s\n", "User ID", session.ID())
			fmt.Fprintf(w, "%s\t%s <%s>\n", "Identity", session.Name(), session.Email())
			fmt.Fprintf(w, "%s\t%s\n", "Username", session.Username())
		}
	}

	fmt.Fprintf(w, " \t \n")
	fmt.Fprintf(w, "%s\t%s\n", "CLI", cfg.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Daemon", daemonVersion.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Registry", registryVersion.Version)
	fmt.Fprintf(w, "%s\t%v\n", "Registry URI", registryURI)
	if loggedIn {
		fmt.Fprintf(w, " \t \n")
		fmt.Fprintf(w, "%s\t%v\n", "Org", ctx.String("org"))
		fmt.Fprintf(w, "%s\t%v\n", "Project", ctx.String("project"))
		fmt.Fprintf(w, "%s\t%v\n", "Environment", ctx.String("environment"))
		fmt.Fprintf(w, "%s\t%v\n", "Service", ctx.String("service"))
		fmt.Fprintf(w, "%s\t%v\n", "Instance", ctx.String("instance"))
	}
	w.Flush()

	return nil
}
