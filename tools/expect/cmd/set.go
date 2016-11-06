package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func set() framework.Command {
	username := "username-" + utils.GlobalNonce
	project := "project-" + utils.GlobalNonce
	// Not very specific environment
	environment := "*"
	service := "service-" + utils.GlobalNonce
	path := fmt.Sprintf("/%s/%s/%s/%s/*/*/secret", username, project, environment, service)
	return framework.Command{
		Spawn: fmt.Sprintf("set %s 1", path),
		Expect: []string{
			"Credentials retrieved",
			"Keypairs retrieved",
			"Encrypting key retrieved",
			"Credential encrypted",
			"Completed Operation",
		},
	}
}

func setOther() framework.Command {
	username := "username-" + utils.GlobalNonce
	project := "project-" + utils.GlobalNonce
	// Not very specific environment
	environment := "*"
	service := "service-" + utils.GlobalNonce
	path := fmt.Sprintf("/%s/%s/%s/%s/*/*/othersecret", username, project, environment, service)
	return framework.Command{
		Spawn: fmt.Sprintf("set %s 1", path),
		Expect: []string{
			"Credentials retrieved",
			"Keypairs retrieved",
			"Encrypting key retrieved",
			"Credential encrypted",
			// This doesn't get output for >1 secret?
			//"Completed Operation",
		},
	}
}

func setSpecific() framework.Command {
	username := "username-" + utils.GlobalNonce
	project := "project-" + utils.GlobalNonce
	// Super specific environment
	environment := "dev-" + username
	service := "service-" + utils.GlobalNonce
	path := fmt.Sprintf("/%s/%s/%s/%s/*/*/secret", username, project, environment, service)
	return framework.Command{
		Spawn: fmt.Sprintf("set %s 2", path),
		Expect: []string{
			"Credentials retrieved",
			"Keypairs retrieved",
			"Encrypting key retrieved",
			"Credential encrypted",
			// This doesn't get output for >1 secret?
			//"Completed Operation",
		},
	}
}
