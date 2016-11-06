package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func view() framework.Command {
	return framework.Command{
		Spawn: fmt.Sprintf("view -s service-%s", utils.GlobalNonce),
		Expect: []string{
			"OTHERSECRET=1",
			"SECRET=1",
		},
	}
}

func viewUnset() framework.Command {
	return framework.Command{
		Spawn: fmt.Sprintf("view -s service-%s", utils.GlobalNonce),
		Expect: []string{
			"SECRET=1",
		},
	}
}

func viewSpecific() framework.Command {
	return framework.Command{
		Spawn: fmt.Sprintf("view -s service-%s", utils.GlobalNonce),
		Expect: []string{
			"SECRET=2",
		},
	}
}
