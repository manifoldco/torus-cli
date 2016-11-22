package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/dirprefs"
	"github.com/manifoldco/torus-cli/errs"
)

func init() {
	unlink := cli.Command{
		Name:     "unlink",
		Usage:    "Remove the link between this project and Torus",
		Category: "PROJECT STRUCTURE",
		Action:   unlinkCmd,
	}

	Cmds = append(Cmds, unlink)
}

func unlinkCmd(ctx *cli.Context) error {
	dPrefs, err := dirprefs.Load(true)
	if err != nil {
		return err
	}

	if dPrefs.Path == "" {
		fmt.Println("No context link exists.")
		return nil
	}

	err = dPrefs.Remove()
	if err != nil {
		return errs.NewErrorExitError("Could not remove link", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if filepath.Dir(dPrefs.Path) == cwd {
		fmt.Println("Your current working directory has been unlinked.")
	} else {
		fmt.Printf("The parent directory '%s' has been unlinked.\n", filepath.Dir(dPrefs.Path))
	}

	return nil
}
