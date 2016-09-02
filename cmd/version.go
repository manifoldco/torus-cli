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
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List version of CLI, Daemon and Registry",
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, listVersionsCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, version)
}

func listVersionsCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	daemonVersion, registryVersion, err := client.Version.Get(c)

	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", "CLI", cfg.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Daemon", daemonVersion.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Registry", registryVersion.Version)
	w.Flush()
	fmt.Println("")

	return nil
}
