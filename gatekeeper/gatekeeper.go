// Package gatekeeper is a web service that will listen for a machine identity from a Cloud
// provider, and request machine credentials.
package gatekeeper

import (
	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/gatekeeper/http"
)

// New returns a new Gatekeeper
func New(org, team string, cfg *config.Config) (g *http.Gatekeeper, err error) {
	api := api.NewClient(cfg)
	http := http.NewGatekeeper(org, team, cfg, api)

	return http, nil
}
