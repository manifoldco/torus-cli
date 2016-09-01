package cmd

import (
	"context"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/promptui"
)

func init() {
	login := cli.Command{
		Name:     "login",
		Usage:    "Log in to your Arigato account",
		Category: "ACCOUNT",
		Action:   Chain(EnsureDaemon, login),
	}
	Cmds = append(Cmds, login)
}

func login(ctx *cli.Context) error {
	prompt := promptui.Prompt{
		Label: "Email",
		Validate: func(input string) error {
			if govalidator.IsEmail(input) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid email address")
		},
	}

	email, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Prompt{
		Label: "Password",
		Mask:  '●',
		Validate: func(input string) error {
			if len(input) > 0 {
				return nil
			}

			return promptui.NewValidationError("Please enter your password")
		},
	}

	password, err := prompt.Run()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	err = client.Session.Login(context.Background(), email, password)
	if err != nil {
		return cli.NewExitError("Login failed. Please try again.", -1)
	}

	fmt.Println("You are now authenticated.")
	return nil
}
