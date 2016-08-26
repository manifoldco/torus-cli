// Package cmd contains all of the Arigato cli commands
package cmd

import (
	"os"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/daemon/config"
)

// Cmds is the list of all cli commands
var Cmds []cli.Command

// loadConfig loads the config, standardizing cli errors on failure.
func loadConfig() (*config.Config, error) {
	arigatoRoot, err := config.CreateArigatoRoot(os.Getenv("ARIGATO_ROOT"))
	if err != nil {
		return nil, cli.NewExitError("Failed to initialize Arigato root dir: "+err.Error(), -1)
	}

	cfg, err := config.NewConfig(arigatoRoot)
	if err != nil {
		return nil, cli.NewExitError("Failed to load config: "+err.Error(), -1)
	}

	return cfg, nil
}
