package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type serviceCreateData struct {
	Name string
}

func serviceCreate() framework.Command {
	service := serviceCreateData{
		Name: "service-" + utils.GlobalNonce,
	}

	return framework.Command{
		Spawn: fmt.Sprintf("services create --org username-%s %s", utils.GlobalNonce, service.Name),
		Expect: []string{
			"Service " + service.Name + " created.",
		},
	}
}

func serviceList() framework.Command {
	return framework.Command{
		Spawn: "services list",
		Expect: []string{
			`project\-` + utils.GlobalNonce + `[\s+]\(1\)`,
			"service-" + utils.GlobalNonce,
		},
	}
}
