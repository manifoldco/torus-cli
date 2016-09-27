package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
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

	password, err := PasswordPrompt(false)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	c := context.Background()
	err = performLogin(c, client, email, password)
	if err != nil {
		return err
	}

	return nil
}

func performLogin(c context.Context, client *api.Client, email, password string) error {
	err := client.Session.Login(context.Background(), email, password)
	if err != nil {
		return cli.NewExitError("Login failed. Please try again.", -1)
	}

	fmt.Println("You are now authenticated.")
	return nil
}
