package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func link() framework.Command {
	return framework.Command{
		Spawn: "link",
		Expect: []string{
			`Org\:[\s]+username-` + utils.GlobalNonce,
			`Project\:[\s]+project-` + utils.GlobalNonce,
			"to view your full working context",
		},
		Prompt: &framework.Prompt{
			Fields: []framework.Field{
				{
					Label:    "Select organization",
					SendLine: "",
				},
				{
					Label:    "Select project",
					SendLine: "",
				},
			},
		},
	}
}

func status() framework.Command {
	org := `username\-` + utils.GlobalNonce
	username := org
	id := org
	project := `project\-` + utils.GlobalNonce
	env := `dev\-` + username
	service := "default"

	return framework.Command{
		Spawn: "status",
		Expect: []string{
			`Identity:[\s]+John Smith`,
			`Username:[\s]+` + username,
			`Org\:[\s]+` + org,
			`Project\:[\s]+` + project,
			`Service\:[\s]+` + service,
			`Instance\:[\s]+1`,
			`Credential path\:[\s]+\/` + org + `\/` + project + `\/` + env + `\/` + service + `\/` + id + `\/1`,
		},
	}
}
