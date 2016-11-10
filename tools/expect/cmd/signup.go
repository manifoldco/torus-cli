package cmd

import (
	"context"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
)

func signup(c *context.Context) framework.Command {
	ctx := framework.ContextValue(c)

	return framework.Command{
		Context: &ctx,
		Spawn:   "signup",
		Expect: []string{
			"Your email is now verified.",
		},
		Timeout: ctx.Timeout,
		Prompt: &framework.Prompt{
			Fields: []framework.Field{
				{
					Label:    "Name",
					SendLine: ctx.User.Name,
				},
				{
					Label:    "Username",
					SendLine: ctx.User.Username,
				},
				{
					Label:    "Email",
					SendLine: ctx.User.Email,
				},
				{
					Label:    "Password",
					SendLine: ctx.User.Password,
				},
				{
					Label:    "Confirm Password",
					SendLine: ctx.User.Password,
					Expect: []string{
						"You are now authenticated.",
						"Keypairs generated",
						"Signing keys signed",
						"Signing keys uploaded",
						"Encryption keys signed",
						"Your account has been created!",
					},
				},
				{
					Label:        "Verification code",
					RequestInput: true,
				},
			},
		},
	}
}
