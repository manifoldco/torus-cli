// Gatekeeper is a machine gateway for automatic machine authentication based on cloud provider
// identity.

package cmd

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/gatekeeper"
)

func init() {
	gatekeeper := cli.Command{
		Name:     "gatekeeper",
		Usage:    "Manage the machine gatekeeper",
		Category: "SYSTEM",
		Subcommands: []cli.Command{
			{
				Name:   "status",
				Usage:  "Display the gatekeeper status",
				Action: gatekeeperStatus,
			},
			{
				Name:  "start",
				Usage: "Start the machine gatekeeper",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "foreground",
						Usage: "Run the gatekeeper in the foreground",
					},
					cli.BoolFlag{
						Name:   "daemonize",
						Usage:  "Background the gatekeeper",
						Hidden: true,
					},
					cli.BoolFlag{
						Name:   "no-permissions-check",
						Usage:  "Skip Torus root dir permission checks",
						Hidden: true,
					},
				},
				Action: chain(ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					func(ctx *cli.Context) error {
						if ctx.Bool("foreground") {
							return startGatekeeperCmd(ctx)
						}
						return spawnGatekeeperCmd()
					},
				),
			},
			{
				Name:   "stop",
				Usage:  "Stop the session daemon",
				Action: stopGatekeeperCmd,
			},
		},
	}

	Cmds = append(Cmds, gatekeeper)
}

// gatekeeperStatus returns the status of the gatekeeper server
func gatekeeperStatus(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil
	}

	return statusListenerCmd("Gatekeeper", cfg.GatekeeperPidPath)
}

// spawnGatekeeper spawns a new process for the Gatekeeper (in the background)
func spawnGatekeeperCmd() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	return spawnListenerCmd("Gatekeeper", cfg.GatekeeperPidPath, gatekeeperCommand)
}

// startGatekeeper starts the machine Gatekeeper
func startGatekeeperCmd(ctx *cli.Context) error {
	return startListenerCmd(
		ctx,
		"Gatekeeper",
		"gatekeeper.log",
		func(cfg *config.Config, noPermissionCheck bool) (Listener, error) {
			return gatekeeper.New(cfg, noPermissionCheck)
		})
}

// stopGatekeeper stops the Gatekeeper process
func stopGatekeeperCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	return stopListenerCmd("Gatekeeper", cfg.GatekeeperPidPath)
}
