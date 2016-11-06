package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type envCreateData struct {
	Name string
}

func envCreate() framework.Command {
	env := envCreateData{
		Name: "env-" + utils.GlobalNonce,
	}

	return framework.Command{
		Spawn: fmt.Sprintf("envs create --org username-%s %s", utils.GlobalNonce, env.Name),
		Expect: []string{
			"Environment " + env.Name + " created.",
		},
	}
}

func envList() framework.Command {
	return framework.Command{
		Spawn: "envs list",
		Expect: []string{
			`project\-` + utils.GlobalNonce + `[\s]+\(2\)`,
			"dev-username-" + utils.GlobalNonce,
			"env-" + utils.GlobalNonce,
		},
	}
}
