package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/kardianos/osext"
	"github.com/natefinch/lumberjack"
	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/daemon"
	"github.com/arigatomachine/cli/daemon/config"
)

func init() {
	daemon := cli.Command{
		Name:  "daemon",
		Usage: "Manage the session daemon",
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "status",
				Usage:  "Display daemon status",
				Action: TODO,
			},
			cli.Command{
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
				},
				Action: func(ctx *cli.Context) error {
					if ctx.Bool("foreground") {
						return startDaemon(ctx)
					}
					return spawnDaemon()
				},
			},
			cli.Command{
				Name:   "stop",
				Usage:  "Stop the session daemon",
				Action: TODO,
			},
		},
	}
	Cmds = append(Cmds, daemon)
}

func spawnDaemon() error {
	executable, err := osext.Executable()
	if err != nil {
		return cli.NewExitError("Unable to find executable: "+err.Error(), -1)
	}

	cmd := exec.Command(executable, "daemon", "start", "--foreground", "--daemonize")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // start a new session group, ie detach
	}

	// Clone the current env, removing email and password if they exist.
	// no need to keep those hanging around in a long lived-process!
	cmd.Env = []string{}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "AG_EMAIL=") || strings.HasPrefix(e, "AG_PASSWORD=") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}

	err = cmd.Start()
	if err != nil {
		return cli.NewExitError("Unable to start daemon: "+err.Error(), -1)
	}

	fmt.Println("Daemon started.")
	return nil
}

func startDaemon(ctx *cli.Context) error {
	arigatoRoot, err := config.CreateArigatoRoot(os.Getenv("ARIGATO_ROOT"))
	if err != nil {
		return cli.NewExitError("Failed to initialize Arigato root dir: "+err.Error(), -1)
	}

	if ctx.Bool("daemonize") {
		log.SetOutput(&lumberjack.Logger{
			Filename:   path.Join(arigatoRoot, "daemon.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})
	}

	cfg, err := config.NewConfig(arigatoRoot)
	if err != nil {
		return cli.NewExitError("Failed to load config: "+err.Error(), -1)
	}

	daemon, err := daemon.New(cfg)
	if err != nil {
		return cli.NewExitError("Failed to create daemon: "+err.Error(), -1)
	}

	go watch(daemon)
	defer daemon.Shutdown()

	log.Printf("v%s of the Daemon is now listening on %s", cfg.Version, daemon.Addr())
	err = daemon.Run()
	if err != nil {
		log.Printf("Error while running daemon: %s", err)
	}

	return err
}

func watch(daemon *daemon.Daemon) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c

	log.Printf("Caught a signal: %s", s)
	shutdown(daemon)
}

func shutdown(daemon *daemon.Daemon) {
	err := daemon.Shutdown()
	if err != nil {
		log.Printf("Did not shutdown cleanly, error: %s", err)
	}

	if r := recover(); r != nil {
		log.Printf("Failed shutting down; caught panic: %v", r)
		panic(r)
	}
}
