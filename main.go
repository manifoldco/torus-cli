//go:generate node cli/passthrough.js
package main

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/kardianos/osext"
	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/cmd"
	"github.com/arigatomachine/cli/config"
)

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		cmd.VersionLookup(ctx)
	}

	app := cli.NewApp()
	app.Version = config.Version
	app.Usage = "A secure, shared workspace for secrets"
	app.Commands = cmd.Cmds
	app.Run(os.Args)
}

func passthrough(ctx *cli.Context, prefixLen int, slug string) error {
	dir, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}

	node := path.Join(dir, "ag-node")
	args := []string{slug}

	for _, f := range ctx.Command.Flags {
		name := strings.SplitN(f.GetName(), ", ", 2)[0]
		switch f.(type) {
		case cli.BoolFlag:
			v := ctx.Bool(name)
			if v {
				args = append(args, "--"+name)
			}
		default:
			v := ctx.String(name)
			if v != "" {
				args = append(args, "--"+name+"="+v)
			}
		}
	}

	args = append(args, ctx.Args()...)

	cmd := exec.Command(node, args...)
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
