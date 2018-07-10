package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func init() {
	signup := cli.Command{
		Name:     "signup",
		Usage:    "Create a new Torus account",
		Category: "ACCOUNT",
		Action:   chain(ensureDaemon, signupCmd),
	}
	Cmds = append(Cmds, signup)
}

func signupCmd(ctx *cli.Context) error {
	return signup(ctx, false)
}

// signup can be ran as a sub-command when an account is needed prior to running
// a particular action. the subCommand boolean signifies it is running as such
// and not as a generic signup
func signup(ctx *cli.Context, subCommand bool) error {
	fmt.Println("Torus is no longer accepting signups.")
	os.Exit(-1)
	return nil
}
