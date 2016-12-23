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
	req, _, err := v.client.NewDaemonRequest("GET", "/version", nil, nil)
	if err != nil {
		return nil, err
	}

	version := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, version)
	return version, err
}
