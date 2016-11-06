package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
)

func logout() framework.Command {
	return framework.Command{
		Spawn: "logout",
		Expect: []string{
			"You have successfully logged out. o/",
		},
	}
}
