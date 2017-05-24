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
)

// Listener is an interface for daemon co-processes that listen on some socket.
type Listener interface {
	// Addr returns the address of the running service. This is a socket in
	// the case of a daemon, or an TCP port in the case of the Gatekeeper.
	Addr() string

	// Run starts the Service. This operation will block.
	Run() error

	// Shutdown shuts the Service, and does any cleanup
	Shutdown() error
}

// spawnListenerCmd is the cli.Command helper to spawn a new Listener
// instance in the background
func spawnListenerCmd(procName, pidPath string, listenerCmd func(string) *exec.Cmd) error {
	proc, err := findListener(pidPath)
	if err != nil {
		return err
	}

	if proc != nil {
		fmt.Printf("%s is already running.\n", procName)
	}

	if err := spawnListener(listenerCmd); err != nil {
		errMsg := fmt.Sprintf("Unsable to start %s", procName)
		return errs.NewErrorExitError(errMsg, err)
	}

	fmt.Printf("%s started.\n", procName)
	return nil
}

// startListenerCmd is the cli.Command helper to start a Listener co-process
func startListenerCmd(
	ctx *cli.Context,
	procName, logPath string,
	listenerFunc func(*config.Config, bool) (Listener, error),
) error {
	noPermissionCheck := ctx.Bool("no-permission-check")
	torusRoot, err := config.CreateTorusRoot(!noPermissionCheck)
	if err != nil {
		return err
	}

	if ctx.Bool("daemonize") {
		log.SetOutput(&lumberjack.Logger{
			Filename:   path.Join(torusRoot, logPath),
			MaxSize:    10, // megabytes
			MaxBackups: 3,  // files
			MaxAge:     28, //days
		})
	} else {
		log.SetOutput(os.Stdout)
	}

	cfg, err := config.NewConfig(torusRoot)
	if err != nil {
		return errs.NewErrorExitError("Failed to load config", err)
	}

	l, err := listenerFunc(cfg, noPermissionCheck)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create %s", procName)
		return errs.NewErrorExitError(errMsg, err)
	}

	go watch(l)
	defer l.Shutdown()

	log.Printf("v%s of the %s is now listening on %s", cfg.Version, procName, l.Addr())

	err = l.Run()
	if err != nil {
		log.Printf("Error while running %s.\n%s", procName, err)
	}

	return err
}

// stopListenerCmd is the cli.Command helper to send the shutdown signal to
// the Listener co-process
func stopListenerCmd(procName, pidPath string) error {
	proc, err := findListener(pidPath)
	if err != nil {
		return err
	}

	if proc == nil {
		fmt.Printf("%s not running\n", procName)
		return nil
	}

	graceful, err := stopProcess(proc)
	if err != nil {
		return err
	}

	if graceful {
		fmt.Printf("%s stopped gracefully.\n", procName)
	} else {
		fmt.Printf("%s stopped forcefully.\n", procName)
	}

	return nil
}

// statusCmd returns the status of the Listener co-process, as well as the Daemon version for API
func statusListenerCmd(procName, pidPath string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	proc, err := findListener(pidPath)
	if err != nil || proc == nil {
		fmt.Printf("%s not running\n", procName)
		return nil
	}

	client := api.NewClient(cfg)
	v, err := client.Version.GetDaemon(context.Background())
	if err != nil {
		return errs.NewErrorExitError("Error communicating with the daemon", err)
	}

	fmt.Printf("%s is running. pid: %d version: v%s\n", procName, proc.Pid, v.Version)

	return nil
}

// watch watches for OS signals to trigger shutdown
func watch(l Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c

	log.Printf("Caught shutdown signal %s", s)
	shutdown(l)
}

// Shutdown shuts the listener down, or panics on failure
func shutdown(l Listener) {
	err := l.Shutdown()
	if err != nil {
		log.Printf("Did not shutdown cleanly.\n%s", err)
	}

	if r := recover(); r != nil {
		log.Printf("Failed shutting down; caught panic.\n%v", r)
		panic(r)
	}
}

// findListener finds and returns the OS Process for the listener. findListener
// returns nil if it is not running, and throws an error if the pid is invalid,
// or there was an issue reading the pid file.
func findListener(pidPath string) (*os.Process, error) {
	lock, err := lockfile.New(pidPath)
	if err != nil {
		return nil, err
	}

	proc, err := lock.GetOwner()
	if err != nil {
		if os.IsNotExist(err) || err == lockfile.ErrInvalidPid || err == lockfile.ErrDeadOwner {
			// cases are OK
			return nil, nil
		}

		return nil, err // we couldn't get the pid, but it should exist, has an owner and is valid
	}

	return proc, nil
}

// findProcess finds the OS process for the corresponding PID
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

// spawnListener spawns a detached Listener instance. Shutdown can be done by
func spawnListener(listenerCmd func(string) *exec.Cmd) error {
	executable, err := osext.Executable()
	if err != nil {
		return errs.NewErrorExitError("Unable to find executable", err)
	}

	cmd := listenerCmd(executable)
	// Clone the current env, removing email and password if they exist.
	// no need to keep those hanging around in a long lived-process!
	cmd.Env = filterEnv()

	return cmd.Start()
}

// stopProcess stops the given OS process. It returns a bool indicating if the
// shutdown was graceful.
func stopProcess(proc *os.Process) (bool, error) {
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
