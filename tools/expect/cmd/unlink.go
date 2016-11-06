package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
)

func unlink() framework.Command {
	return framework.Command{
		Spawn: "unlink",
		Expect: []string{
			"Your current working directory has been unlinked.",
		},
	}
}
