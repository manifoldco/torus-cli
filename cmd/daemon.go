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
	"strings"
	"syscall"
	"time"

	"github.com/kardianos/osext"
	"github.com/natefinch/lumberjack"
	"github.com/nightlyone/lockfile"
	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"

	"github.com/arigatomachine/cli/daemon"
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
	v, err := client.Version.Get(context.Background())
	if err != nil {
		return cli.NewExitError("Error communicating with the daemon: "+err.Error(), -1)
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
		if strings.HasPrefix(e, "TORUS_EMAIL=") || strings.HasPrefix(e, "TORUS_PASSWORD=") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}

	err = cmd.Start()
	if err != nil {
		return cli.NewExitError("Unable to start daemon: "+err.Error(), -1)
	}

	return nil
}

func startDaemon(ctx *cli.Context) error {
	torusRoot, err := config.CreateTorusRoot()
	if err != nil {
		return cli.NewExitError("Failed to initialize Torus root dir: "+err.Error(), -1)
	}

	if ctx.Bool("daemonize") {
		log.SetOutput(&lumberjack.Logger{
			Filename:   path.Join(torusRoot, "daemon.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})
	}

	cfg, err := config.NewConfig(torusRoot)
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

	increment := 50 * time.Millisecond
	for d := increment; d < 3*time.Second; d += increment {
		time.Sleep(d)
		if _, err := findProcess(proc.Pid); err != nil {
			return true, nil
		}
	}

	err = proc.Kill()
	if err != nil {
		return false, cli.NewExitError("Could not stop daemon: "+err.Error(), -1)
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
