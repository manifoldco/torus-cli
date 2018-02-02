package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/pathexp"

	"github.com/urfave/cli"
)

func init() {
	run := cli.Command{
		Name:      "run",
		Usage:     "Run a process and inject secrets into its environment",
		ArgsUsage: "[--] <command> [<arguments>...]",
		Category:  "SECRETS",
		Flags: []cli.Flag{
			stdOrgFlag,
			stdProjectFlag,
			stdEnvFlag,
			serviceFlag("Use this service.", "default", true),
		},
		Action: chain(
			ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
			setUserEnv, checkRequiredFlags, runCmd,
		),
	}

	Cmds = append(Cmds, run)
}

func runCmd(ctx *cli.Context) error {
	args := ctx.Args()

	if len(args) == 0 {
		return errs.NewUsageExitError("A command is required", ctx)
	} else if len(args) == 1 { // only one arg? maybe it was quoted
		args = strings.Split(args[0], " ")
	}

	secrets, path, err := getSecrets(ctx)
	if err != nil {
		return err
	}

	// Create the command. It gets this processes's stdio.
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = manipulateEnv(path)

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

// manipulateEnv removes any sensitive torus environment variables and sets
// `TORUS_ORG`, `TORUS_PROJECT`, `TORUS_ENVIRONMENT`, and `TORUS_SERVICE` for
// use by the running process.
func manipulateEnv(path *pathexp.PathExp) []string {
	env := filterEnv()

	org := path.Org.Components()
	project := path.Project.Components()
	envs := path.Envs.Components()
	services := path.Services.Components()

	env = append(env, "TORUS_ORG="+org[0])
	env = append(env, "TORUS_PROJECT="+project[0])
	env = append(env, "TORUS_ENVIRONMENT="+envs[0])
	env = append(env, "TORUS_SERVICE="+services[0])

	return env
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
