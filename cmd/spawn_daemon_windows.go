package cmd

import (
	"os/exec"
	"syscall"
)

func daemonCommand(executable string) *exec.Cmd {
	cmd := exec.Command(executable, "daemon", "start", "--foreground", "--daemonize")
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	return cmd
}
