package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type orgCreateData struct {
	Name string
}

func orgCreate() framework.Command {
	org := orgCreateData{
		Name: "new-org-" + utils.GlobalNonce,
	}

	return framework.Command{
		Spawn: fmt.Sprintf("orgs create %s", org.Name),
		Expect: []string{
			"Keypairs generated",
			"Signing keys signed",
			"Signing keys uploaded",
			"Encryption keys signed",
			"Org " + org.Name + " created.",
		},
	}
}

func orgList() framework.Command {
	return framework.Command{
		Spawn: "orgs list",
		Expect: []string{
			`username\-` + utils.GlobalNonce + `\s+\[personal\]`,
			`new\-org\-` + utils.GlobalNonce,
		},
	}
}
