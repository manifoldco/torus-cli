package routes

import (
	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
)

// NewRoutesMux returns a new *box.Mux for handling Gatekeeper requests
func NewRoutesMux(cfg *config.Config, api *api.Client) *bone.Mux {
	mux := bone.New()

	mux.PostFunc("/machine", machineCreateRoute(api))

	return mux
}
