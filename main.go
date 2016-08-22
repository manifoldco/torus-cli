//go:generate node cli/passthrough.js
package main

import (
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/kardianos/osext"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "A secure, shared workspace for secrets"
	app.Commands = passthroughs
	app.Run(os.Args)
}

func passthrough(ctx *cli.Context, prefixLen int, slug string) error {
	dir, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}

	node := path.Join(dir, "ag-node")
	cmd := exec.Command(node, append([]string{slug}, os.Args[prefixLen:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
			return nil
		}
	}
	return err
}
