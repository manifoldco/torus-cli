package cmd

import (
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/daemon"
)

func init() {
	daemon := cli.Command{
		Name:     "daemon",
		Usage:    "Manage the session daemon",
		Category: "SYSTEM",
		Subcommands: []cli.Command{
			{
				Name:   "status",
				Usage:  "Display daemon status",
				Action: daemonStatus,
			},
			{
				Name:  "start",
				Usage: "Start the session daemon",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "foreground",
						Usage: "Run the Daemon in the foreground",
					},
					cli.BoolFlag{
						Name:   "daemonize",
						Usage:  "Run as a background session daemon",
						Hidden: true, // not displayed in help; used internally
					},
					cli.BoolFlag{
						Name:   "no-permission-check",
						Usage:  "Skip Torus root dir permission checks",
						Hidden: true, // Just for system daemon use
					},
				},
				Action: func(ctx *cli.Context) error {
					if ctx.Bool("foreground") {
						return startDaemon(ctx)
					}
					return spawnDaemonCmd()
				},
			},
			{
				Name:   "stop",
				Usage:  "Stop the session daemon",
				Action: stopDaemonCmd,
			},
		},
	}
	Cmds = append(Cmds, daemon)
}

func daemonStatus(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	return statusListenerCmd("Daemon", cfg.PidPath)
}

func spawnDaemonCmd() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	return spawnListenerCmd("Daemon", cfg.PidPath, daemonCommand)
}

func startDaemon(ctx *cli.Context) error {
	return startListenerCmd(
		ctx,
		"Daemon",
		"daemon.log",
		func(cfg *config.Config, noPermissionCheck bool) (Listener, error) {
			return daemon.New(cfg, noPermissionCheck)
		})
}

func stopDaemonCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	return stopListenerCmd("Daemon", cfg.PidPath)
}
