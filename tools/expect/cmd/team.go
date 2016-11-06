package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type teamCreateData struct {
	Name string
}

func teamCreate() framework.Command {
	team := teamCreateData{
		Name: "team-" + utils.GlobalNonce,
	}

	return framework.Command{
		Spawn: fmt.Sprintf("teams create %s", team.Name),
		Expect: []string{
			"Team " + team.Name + " created.",
		},
	}
}

func teamList() framework.Command {
	return framework.Command{
		Spawn: "teams list",
		Expect: []string{
			`\* owner[\s]+\[system\]`,
			`\* admin[\s]+\[system\]`,
			`\* member[\s]+\[system\]`,
			"team-" + utils.GlobalNonce,
		},
	}
}
