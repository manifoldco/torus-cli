package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/config"
)

func init() {
	logout := cli.Command{
		Name:     "logout",
		Usage:    "Log out of your Torus account",
		Category: "ACCOUNT",
		Action:   chain(ensureDaemon, logoutCmd),
	}
	Cmds = append(Cmds, logout)
}

func logoutCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	err = client.Session.Logout(context.Background())
	if err != nil {
		if herr, ok := err.(*apitypes.Error); ok {
			if herr.StatusCode == 404 {
				fmt.Println("You are not logged in.")
				return nil
			}
		}
		return cli.NewExitError("Logout failed. Please try again.", -1)
	}

	fmt.Println("You have successfully logged out. o/")
	return nil
}
