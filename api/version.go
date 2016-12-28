package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/registry"
)

// VersionClient provides access to the daemon's /v1/version endpoint, for
// inspecting the daemon's release version.
type VersionClient struct {
	*registry.VersionClient
	client *apiRoundTripper
}

func newVersionClient(upstream *registry.VersionClient, rt *apiRoundTripper) *VersionClient {
	return &VersionClient{upstream, rt}
}

// GetDaemon returns the daemon's release version.
func (v *VersionClient) GetDaemon(ctx context.Context) (*apitypes.Version, error) {
	version := &apitypes.Version{}
	err := v.client.DaemonRoundTrip(ctx, "GET", "/version", nil, nil, version, nil)
	return version, err
}
