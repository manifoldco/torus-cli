package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
)

// VersionClient provides access to the daemon's /v1/version endpoint, for
// inspecting the daemon's release version.
type VersionClient struct {
	client *Client
}

// Get returns the daemon's release version.
func (v *VersionClient) Get(ctx context.Context) (*apitypes.Version, error) {
	return v.fetchVersion(ctx, false)
}

// GetRegistry returns the registry's release version.
func (v *VersionClient) GetRegistry(ctx context.Context) (*apitypes.Version, error) {
	return v.fetchVersion(ctx, true)
}

func (v *VersionClient) fetchVersion(ctx context.Context, proxied bool) (*apitypes.Version, error) {
	req, _, err := v.client.NewRequest("GET", "/version", nil, nil, proxied)
	if err != nil {
		return nil, err
	}

	version := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, version, nil, nil)
	return version, err
}
