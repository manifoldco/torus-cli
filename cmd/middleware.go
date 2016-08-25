package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/apitypes"
)

// Chain allows easy sequential calling of BeforeFuncs and AfterFuncs.
// Chain will exit on the first error seen.
// XXX Chain is only public while we need it for passthrough.go
func Chain(funcs ...func(*cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {

		for _, f := range funcs {
			err := f(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// EnsureDaemon ensures that the daemon is running, and is the correct version,
// before a command is exeucted.
// the daemon will be started/restarted once, to try and launch the latest
// version.
// XXX EnsureDaemon is only public while we need it for passthrough.go
func EnsureDaemon(ctx *cli.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	proc, err := findDaemon(cfg)
	if err != nil {
		return err
	}

	spawned := false

	if proc == nil {
		err := spawnDaemon()
		if err != nil {
			return err
		}

		spawned = true
	}

	client := api.NewClient(cfg)

	var v *apitypes.Version
	increment := 5 * time.Millisecond
	for d := increment; d < 1*time.Second; d += increment {
		v, err = client.Version.Get(context.Background())
		if err == nil {
			break
		}
		time.Sleep(d)
	}

	if err != nil {
		return cli.NewExitError("Error communicating with the daemon: "+err.Error(), -1)
	}

	if v.Version == cfg.Version {
		return nil
	}

	if spawned {
		return cli.NewExitError("The daemon version is incorrect. Check for stale processes", -1)
	}

	fmt.Println("The daemon version is out of date and is being restarted.")
	fmt.Println("You will need to login again.")

	_, err = stopDaemon(proc)
	if err != nil {
		return err
	}

	return EnsureDaemon(ctx)
}
