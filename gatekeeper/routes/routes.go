package routes

import (
	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/api"
)

// NewRoutesMux returns a new *box.Mux for handling Gatekeeper requests
func NewRoutesMux(org, team string, api *api.Client) *bone.Mux {
	mux := bone.New()

	mux.SubRoute("/machine", newMachineBootstrapRoutes(org, team, api))

	return mux
}

func newMachineBootstrapRoutes(org, team string, api *api.Client) *bone.Mux {
	mux := bone.New()

	mux.PostFunc("/aws", awsBootstrapRoute(org, team, api))

	return mux
}
