package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/kardianos/osext"
	"github.com/natefinch/lumberjack"
	"github.com/nightlyone/lockfile"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"

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

	proc, err := findDaemon(cfg)
	if err != nil {
		return err
	}

	if proc == nil {
		fmt.Println("Daemon is not running.")
		return nil
	}

	client := api.NewClient(cfg)
	v, err := client.Version.GetDaemon(context.Background())
	if err != nil {
		return errs.NewErrorExitError("Error communicating with the daemon", err)
	}

	fmt.Printf("Daemon is running. pid: %d version: v%s\n", proc.Pid, v.Version)

	return nil
}

func spawnDaemonCmd() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	proc, err := findDaemon(cfg)
	if err != nil {
		return err
	}

	if proc != nil {
		fmt.Println("Daemon is already running.")
		return nil
	}

	err = spawnDaemon()
	if err != nil {
		return err
	}

	fmt.Println("Daemon started.")
	return nil
}

func spawnDaemon() error {
	executable, err := osext.Executable()
	if err != nil {
		return errs.NewErrorExitError("Unable to find executable.", err)
	}

	cmd := exec.Command(executable, "daemon", "start", "--foreground", "--daemonize")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // start a new session group, ie detach
	}

	// Clone the current env, removing email and password if they exist.
	// no need to keep those hanging around in a long lived-process!
	cmd.Env = filterEnv()

	err = cmd.Start()
	if err != nil {
		return errs.NewErrorExitError("Unable to start daemon", err)
	}

	return nil
}

func startDaemon(ctx *cli.Context) error {
	noPermissionCheck := ctx.Bool("no-permission-check")
	torusRoot, err := config.CreateTorusRoot(!noPermissionCheck)
	if err != nil {
		return errs.NewErrorExitError("Failed to initialize Torus root dir.", err)
	}

	if ctx.Bool("daemonize") {
		log.SetOutput(&lumberjack.Logger{
			Filename:   path.Join(torusRoot, "daemon.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})
	} else {
		// re-enable logging, as by default its silenced for foreground use.
		log.SetOutput(os.Stdout)
	}

	cfg, err := config.NewConfig(torusRoot)
	if err != nil {
		return errs.NewErrorExitError("Failed to load config.", err)
	}

	daemon, err := daemon.New(cfg, noPermissionCheck)
	if err != nil {
		return errs.NewErrorExitError("Failed to create daemon.", err)
	}

	go watch(daemon)
	defer daemon.Shutdown()

	log.Printf("v%s of the Daemon is now listening on %s", cfg.Version, daemon.Addr())
	err = daemon.Run()
	if err != nil {
		log.Printf("Error while running daemon.\n%s", err)
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
		log.Printf("Did not shutdown cleanly.\n%s", err)
	}

	if r := recover(); r != nil {
		log.Printf("Failed shutting down; caught panic.\n%v", r)
		panic(r)
	}
}

func stopDaemonCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	proc, err := findDaemon(cfg)
	if err != nil {
		return err
	}

	if proc == nil {
		fmt.Println("Daemon is not running.")
		return nil
	}

	graceful, err := stopDaemon(proc)
	if err != nil {
		return err
	}

	if graceful {
		fmt.Println("Daemon stopped gracefully.")
	} else {

		fmt.Println("Daemon stopped forcefully.")
	}

	return nil
}

// stopDaemon stops the daemon process. It returns a bool indicating if the
// shutdown was graceful.
func stopDaemon(proc *os.Process) (bool, error) {
	err := proc.Signal(syscall.Signal(syscall.SIGTERM))

	if err == nil { // no need to wait for the SIGTERM if the signal failed
		increment := 50 * time.Millisecond
		for d := increment; d < 3*time.Second; d += increment {
			time.Sleep(d)
			if _, err := findProcess(proc.Pid); err != nil {
				return true, nil
			}
		}
	}

	err = proc.Kill()
	if err != nil {
		return false, errs.NewErrorExitError("Could not stop daemon.", err)
	}

	return false, nil
}

// findDaemon returns an os.Process for a running daemon, or nil if it is not
// running. It returns an error if the pid file location is invalid in the
// config, or there was an error reading the pid file.
func findDaemon(cfg *config.Config) (*os.Process, error) {
	lock, err := lockfile.New(cfg.PidPath)
	if err != nil {
		return nil, err
	}

	proc, err := lock.GetOwner()
	if err != nil {
		// happy path cases. the pid doesn't exist, or it contains garbage,
		// or the previous owner is gone.
		if os.IsNotExist(err) || err == lockfile.ErrInvalidPid || err == lockfile.ErrDeadOwner {
			return nil, nil
		}

		return nil, err
	}

	return proc, nil
}

func findProcess(pid int) (*os.Process, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	// On unix, findprocess does not error. So we have to check for the process.
	if runtime.GOOS != "windows" {
		err = proc.Signal(syscall.Signal(0))
		if err != nil {
			return nil, err
		}
	}

	return proc, nil
}
