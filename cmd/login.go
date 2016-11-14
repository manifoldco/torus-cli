package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
)

func init() {
	login := cli.Command{
		Name:     "login",
		Usage:    "Log in to your Torus account",
		Category: "ACCOUNT",
		Action:   chain(ensureDaemon, login),
	}
	Cmds = append(Cmds, login)
}

func login(ctx *cli.Context) error {
	email, err := EmailPrompt("")
	if err != nil {
		return err
	}

	password, err := PasswordPrompt(false, nil)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	c := context.Background()
	err = performLogin(c, client, email, password, true)
	if err != nil {
		return err
	}

	return nil
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
