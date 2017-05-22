package routes

import (
	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/api"
)

// NewRoutesMux returns a new *box.Mux for handling Gatekeeper requests
func NewRoutesMux(api *api.Client) *bone.Mux {
	mux := bone.New()

	mux.SubRoute("/machine", newMachineBootstrapRoutes(api))

	return mux
}

func newMachineBootstrapRoutes(api *api.Client) *bone.Mux {
	mux := bone.New()

	mux.PostFunc("/aws", awsBootstrapRoute(api))

	return mux
}
