package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type loginData struct {
	Email    string
	Password string
}

func login() framework.Command {
	user := loginData{
		Email:    "email+" + utils.GlobalNonce + "@arigato.sh",
		Password: "password",
	}
	timeout := utils.Duration(10)

	return framework.Command{
		Spawn: "login",
		Expect: []string{
			"You are now authenticated.",
		},
		Timeout: &timeout,
		Prompt: &framework.Prompt{
			Fields: []framework.Field{
				{
					Label:    "Email",
					SendLine: user.Email,
				},
				{
					Label:    "Password",
					SendLine: user.Password,
				},
			},
		},
	}
}
