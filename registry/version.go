package registry

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
)

// VersionClient provides access to the daemon's /v1/version endpoint, for
// inspecting the daemon's release version.
type VersionClient struct {
	client RoundTripper
}

// Get returns the registry's release version.
func (v *VersionClient) Get(ctx context.Context) (*apitypes.Version, error) {
	version := apitypes.Version{}
	err := v.client.RoundTrip(ctx, "GET", "/version", nil, nil, &version)
	return &version, err
}
