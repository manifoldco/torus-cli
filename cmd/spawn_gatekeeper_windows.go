package cmd

import (
	"os/exec"
	"syscall"
)

func gatekeeperCommand(executable string) *exec.Cmd {
	cmd := exec.Command(executable, "gatekeeper", "start", "--foreground", "--daemonize")
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	return cmd
}
