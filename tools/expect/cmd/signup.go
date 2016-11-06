package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type signupData struct {
	Name     string
	Email    string
	Username string
	Password string
}

func signup() framework.Command {
	if utils.GlobalNonce == "" {
		panic("empty nonce")
	}
	user := signupData{
		Name:     "John Smith",
		Username: "username-" + utils.GlobalNonce,
		Email:    "email+" + utils.GlobalNonce + "@arigato.sh",
		Password: "password",
	}
	timeout := utils.Duration(30)

	return framework.Command{
		Spawn: "signup",
		Expect: []string{
			"Your email is now verified.",
		},
		Timeout: &timeout,
		Prompt: &framework.Prompt{
			Fields: []framework.Field{
				{
					Label:    "Name",
					SendLine: user.Name,
				},
				{
					Label:    "Username",
					SendLine: user.Username,
				},
				{
					Label:    "Email",
					SendLine: user.Email,
				},
				{
					Label:    "Password",
					SendLine: user.Password,
				},
				{
					Label:    "Confirm Password",
					SendLine: user.Password,
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
