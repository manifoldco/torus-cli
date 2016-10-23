// Package cmd contains all of the Torus cli commands
package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
)

// Cmds is the list of all cli commands
var Cmds []cli.Command

var progress api.ProgressFunc = func(evt *api.Event, err error) {
	if evt != nil {
		fmt.Println(evt.Message)
	}
}
