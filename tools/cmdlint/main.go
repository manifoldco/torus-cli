package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/cmd"
)

func lintCmd(c cli.Command) error {
	var err error
	order := []string{"create", "list", "view", "delete"}
	var neworder []string

orderLoop:
	for _, o := range order {
		for _, s := range c.Subcommands {
			if s.Names()[0] == o {
				neworder = append(neworder, o)
				continue orderLoop
			}
		}
	}

	for _, s := range c.Subcommands {
		name := s.Names()[0]
		if len(neworder) == 0 {
			break
		}
		if neworder[0] == name {
			neworder = neworder[1:]
		}

		for _, o := range neworder {
			if o == name {
				fmt.Printf("Error: %s standard subcommand '%s' out of order\n",
					c.FullName(), name)
				err = errors.New("subcommand out of order")
			}
		}
	}

	for _, s := range c.Subcommands {
		newerr := lintCmd(s)
		if newerr != nil {
			err = newerr
		}
	}

	return err
}

func main() {
	var err error
	for _, c := range cmd.Cmds {
		newerr := lintCmd(c)
		if newerr != nil {
			err = newerr
		}
	}

	if err != nil {
		os.Exit(1)
	}
}
