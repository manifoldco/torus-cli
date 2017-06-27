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

var (
	certFlag = newPlaceholder("cert, c", "CERT", "Certificate for SSL", "", "TORUS_GATEKEEPER_CERT", false)
	keyFlag  = newPlaceholder("key, k", "KEY", "Certificate key for SSL", "", "TORUS_GATEKEEPER_CERT_KEY", false)
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
				Action: chain(ensureDaemon, ensureSession, loadDirPrefs,
					loadPrefDefaults, startGatekeeperCmd,
				),
				Flags: []cli.Flag{
					orgFlag("Use this organization by default in Gatekeeper", false),
					roleFlag("Use this role.", false),
					certFlag,
					keyFlag,
				},
			},
		},
	}

	Cmds = append(Cmds, gatekeeper)
}

// startGatekeeper starts the machine Gatekeeper
func startGatekeeperCmd(ctx *cli.Context) error {
	log.SetOutput(os.Stdout)

	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewErrorExitError("Failed to load config.", err)
	}

	gatekeeper, err := gatekeeper.New(ctx.String("org"), ctx.String("role"), ctx.String("cert"), ctx.String("key"), cfg)
	if err != nil {
		log.Printf("Error starting a new Gatekeeper instance: %s", err)
		return err
	}

	log.Printf("v%s of the Gatekeeper is now listeneing on %s", cfg.Version, gatekeeper.Addr())
	err = gatekeeper.Listen()
	if err != nil {
		log.Printf("Error while running the Gatekeeper.\n%s", err)
	}

	return err
}
