package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
)

func prefsList() framework.Command {
	return framework.Command{
		Spawn: "prefs list",
		Expect: []string{
			// Only run expect tests against local
			"registry_uri    = http://localhost:8080",
		},
	}
}
