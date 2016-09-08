package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/promptui"
)

const acceptInviteFailed = "Could not accept invitation to org, please try again."

func invitesAccept(ctx *cli.Context) error {
	usage := usageString(ctx)

	args := ctx.Args()
	if len(args) < 2 {
		var text string
		if len(args) < 1 {
			text = "Missing email and code"
		} else {
			text = "Missing code"
		}
		text = text + "\n\n"
		text += usage
		return cli.NewExitError(text, -1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return cli.NewExitError(envCreateFailed, -1)
	}

	client := api.NewClient(cfg)
	c := context.Background()

	_, err = client.Session.Get(c)
	if err != nil {
		_, value, err := SelectAcceptAction()
		if err != nil {
			if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
				return err
			}
			return cli.NewExitError(acceptInviteFailed, -1)
		}

		// Become logged in either through signup or login
		switch value {
		case "Login":
			fmt.Println("")
			err = login(ctx)
			if err != nil {
				return err
			}
		case "Signup":
			fmt.Println("")
			err = signup(ctx, true)
			if err != nil {
				return err
			}
		default:
			return cli.NewExitError(acceptInviteFailed, -1)
		}
		fmt.Println("")
	}

	email := args[0]
	code := args[1]
	err = validateInviteCode(code)
	if err != nil {
		return err
	}

	invite, err := client.Invites.Associate(c, ctx.String("org"), email, code)
	if err != nil || invite == nil {
		return cli.NewExitError(acceptInviteFailed, -1)
	}

	err = generateKeypairsForOrg(ctx, c, client, invite.Body.OrgID, false)
	if err != nil {
		// We'd rather they generate keypairs through accept again, so generic err
		return cli.NewExitError(acceptInviteFailed, -1)
	}

	err = client.Invites.Accept(c, ctx.String("org"), email, code)
	if err != nil {
		return cli.NewExitError(acceptInviteFailed, -1)
	}

	fmt.Println("You have accepted the invitation.")
	fmt.Println("\nYou will be added to the org once the administrator has approved your invite.")
	return nil
}
