package routes

import (
	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/api"
)

// NewRoutesMux returns a new *box.Mux for handling Gatekeeper requests
func NewRoutesMux(api *api.Client) *bone.Mux {
	mux := bone.New()

	mux.PostFunc("/machine", machineCreateRoute(api))

	return mux
}
