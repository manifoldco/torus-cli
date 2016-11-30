// Package cmd contains all of the Torus cli commands
package cmd

import (
	"context"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/ui"
)

// Cmds is the list of all cli commands
var Cmds []cli.Command

var progress api.ProgressFunc = func(evt *api.Event, err error) {
	if evt != nil {
		ui.Progress(evt.Message)
	}
}

// NewAPIClient loads config and creates a new api client
func NewAPIClient(ctx *context.Context, client *api.Client) (context.Context, *api.Client, error) {
	if client == nil {
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, nil, err
		}
		client = api.NewClient(cfg)
	}
	var c context.Context
	if ctx == nil {
		c = context.Background()
	} else {
		c = *ctx
	}
	return c, client, nil
}
