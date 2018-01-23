package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prompts"
)

func init() {
	login := cli.Command{
		Name:     "login",
		Usage:    "Log in to a user account, authenticating the CLI",
		Category: "ACCOUNT",
		Action:   chain(ensureDaemon, login, checkUpdates),
	}
	Cmds = append(Cmds, login)
}

func login(ctx *cli.Context) error {
	email, err := prompts.Email("", false)
	if err != nil {
		return err
	}

	password, err := prompts.Password(false, nil)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	c := context.Background()
	return performLogin(c, client, email, password, true)
}

func performLogin(c context.Context, client *api.Client, email, password string, shouldPrint bool) error {
	err := client.Session.UserLogin(context.Background(), email, password)
	if err != nil {
		return errs.NewErrorExitError("Login failed.", err)
	}

	if shouldPrint {
		fmt.Println("You are now authenticated.")
	}
	return nil
}

func testLogin(c context.Context, client *api.Client, email, password string) error {
	return performLogin(c, client, email, password, false)
}
