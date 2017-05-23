// Gatekeeper is a machine gateway for automatic machine authentication based on cloud provider
// identity.

package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/gatekeeper"
)

func init() {
	gatekeeper := cli.Command{
		Name:     "gatekeeper",
		Usage:    "Manage the machine gatekeeper",
		Category: "SYSTEM",
		Subcommands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start the machine gatekeeper",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:   "no-permission-check",
						Usage:  "Skip Torus root dir permission checks",
						Hidden: true, // Just for system daemon use
					},
				},
				Action: chain(ensureDaemon, ensureSession, loadDirPrefs,
					loadPrefDefaults, startGatekeeperCmd,
				),
			},
		},
	}

	Cmds = append(Cmds, gatekeeper)
}

// startGatekeeper starts the machine Gatekeeper
func startGatekeeperCmd(ctx *cli.Context) error {
	noPermissionCheck := ctx.Bool("no-permission-check")
	torusRoot, err := config.CreateTorusRoot(!noPermissionCheck)
	if err != nil {
		return errs.NewErrorExitError("Failed to initialize Torus root dir.", err)
	}

	log.SetOutput(os.Stdout)

	cfg, err := config.NewConfig(torusRoot)
	if err != nil {
		return errs.NewErrorExitError("Failed to load config.", err)
	}

	gatekeeper, err := gatekeeper.New(ctx, cfg)

	log.Printf("v%s of the Gatekeeper is now listeneing on %s", cfg.Version, gatekeeper.Addr())
	err = gatekeeper.Run()
	if err != nil {
		log.Printf("Error while running the Gatekeeper.\n%s", err)
	}

	return err
}
