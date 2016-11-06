package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func version() framework.Command {
	return framework.Command{
		Spawn: "version",
		Expect: []string{
			`CLI[\s]+` + utils.SemverRegex,
			`Daemon[\s]+` + utils.SemverRegex,
			`Registry[\s]+` + utils.SemverRegex,
		},
	}
}
