package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"github.com/manifoldco/torus-cli/errs"

	"github.com/urfave/cli"
	"fmt"
)

func init() {
	run := cli.Command{
		Name:      "run",
		Usage:     "Run a process and inject secrets into its environment",
		ArgsUsage: "[--] <command> [<arguments>...]",
		Category:  "SECRETS",
		Flags: []cli.Flag {
			stdOrgFlag,
			stdProjectFlag,
			stdEnvFlag,
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "Show credential path.",
			},
			userFlag("Use this user.", false),
			machineFlag("Use this machine.", false),
			serviceFlag("Use this service.", "default", true),
			stdInstanceFlag,
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, checkRequiredFlags, runCmd,
		),
	}

	Cmds = append(Cmds, run)
}

func runCmd(ctx *cli.Context) error {
	fmt.Printf("runCmd")
	args := ctx.Args()
	shouldShowCreds := ctx.Bool("verbose")
	if len(args) == 0 {
		return errs.NewUsageExitError("A command is required", ctx)
	} else if len(args) == 1 { // only one arg? maybe it was quoted
		args = strings.Split(args[0], " ")
	}

	secrets, path, err := getSecrets(ctx)
	if err != nil {
		return err
	}
	if shouldShowCreds {
		fmt.Printf(path)
	}

	// Create the command. It gets this processes's stdio.
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = filterEnv()

	// Add the secrets into the env
	for _, secret := range secrets {
		value := (*secret.Body).GetValue()
		key := strings.ToUpper((*secret.Body).GetName())

		cmd.Env = append(cmd.Env, key+"="+value.String())
	}

	err = cmd.Start()
	if err != nil {
		return errs.NewErrorExitError("Failed to run command", err)
	}

	done := make(chan bool)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c) // give us all signals to relay

		select {
		case s := <-c:
			cmd.Process.Signal(s)
		case <-done:
			signal.Stop(c)
			return
		}
	}()

	err = cmd.Wait()
	close(done)
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
				return nil
			}
		}
		return err
	}

	return nil
}

func filterEnv() []string {
	env := []string{}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "TORUS_EMAIL=") || strings.HasPrefix(e, "TORUS_PASSWORD=") ||
			strings.HasPrefix(e, "TORUS_TOKEN_ID=") || strings.HasPrefix(e, "TORUS_TOKEN_SECRET=") {
			continue
		}
		env = append(env, e)
	}

	return env
}
