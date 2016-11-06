package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type projectCreateData struct {
	Name string
}

func projectCreate() framework.Command {
	project := projectCreateData{
		Name: "project-" + utils.GlobalNonce,
	}

	return framework.Command{
		Spawn: fmt.Sprintf("projects create --org username-%s %s", utils.GlobalNonce, project.Name),
		Expect: []string{
			"Project " + project.Name + " created.",
		},
	}
}

func projectList() framework.Command {
	return framework.Command{
		Spawn: "projects list",
		Expect: []string{
			`username\-` + utils.GlobalNonce + `[\s+]org[\s+]\(1\)`,
			"project-" + utils.GlobalNonce,
		},
	}
}
