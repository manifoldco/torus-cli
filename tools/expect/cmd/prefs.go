package cmd

import (
	"context"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
)

func prefsList(c *context.Context) framework.Command {
	ctx := framework.ContextValue(c)

	return framework.Command{
		Context: &ctx,
		Spawn:   "prefs list",
		Expect: []string{
			// Only run expect tests against local
			"registry_uri    = http://localhost:8080",
		},
	}
}
