// Package gatekeeper is a web service that will listen for a machine identity from a Cloud
// provider, and request machine credentials.
package gatekeeper

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/gatekeeper/http"
)

// Gatekeeper provides the Listener interface for a daemonized HTTP co-process that will act
// to register machines
type Gatekeeper struct {
	g *http.Gatekeeper
}

// New returns a new Gatekeeper
func New(ctx *cli.Context, cfg *config.Config) (g *Gatekeeper, err error) {
	// recovery and cleanup
	defer func() {
		if r := recover(); r != nil {
			// named return
			g = nil
			err = r.(error)
		}
	}()

	api := api.NewClient(cfg)
	http := http.NewGatekeeper(ctx, cfg, api)

	gatekeeper := &Gatekeeper{
		g: http,
	}

	return gatekeeper, nil
}

// Addr returns the address of the currently running Gatekeeper
func (g *Gatekeeper) Addr() string {
	return g.g.Addr()
}

// Run starts the Gatekeeper main loop. If blocks until the service is shutdown, or encounters an
// error
func (g *Gatekeeper) Run() error {
	return g.g.Listen()
}

// Shutdown gracefully stops the HTTP service
func (g *Gatekeeper) Shutdown() error {
	if err := g.g.Close(); err != nil {
		return fmt.Errorf("Could not stop gatekeeper: %s", err)
	}

	return nil
}
