package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
)

func init() {
	version := cli.Command{
		Name:     "version",
		Usage:    "Display versions of utility components",
		Category: "SYSTEM",
		Action:   VersionLookup,
	}
	Cmds = append(Cmds, version)
}

// VersionLookup ensures the environment is ready and then executes version cmd
func VersionLookup(ctx *cli.Context) error {
	return Chain(
		EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
		SetUserEnv, checkRequiredFlags, listVersionsCmd,
	)(ctx)
}

func listVersionsCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	daemonVersion, registryVersion, err := client.Version.Get(c)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", "CLI", cfg.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Daemon", daemonVersion.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Registry", registryVersion.Version)
	w.Flush()

	return nil
}
