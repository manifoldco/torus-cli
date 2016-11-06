package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func unsetOther() framework.Command {
	username := "username-" + utils.GlobalNonce
	project := "project-" + utils.GlobalNonce
	// Not very specific environment
	environment := "*"
	service := "service-" + utils.GlobalNonce
	path := fmt.Sprintf("/%s/%s/%s/%s/*/*/othersecret", username, project, environment, service)
	return framework.Command{
		Spawn: fmt.Sprintf("unset %s -y", path),
		Expect: []string{
			"Credentials retrieved",
			"Keypairs retrieved",
			"Encrypting key retrieved",
			"Credential encrypted",
			"Completed Operation",
		},
	}
}
