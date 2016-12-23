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
	req, err := v.client.NewRequest("GET", "/version", nil, nil)
	if err != nil {
		return nil, err
	}

	version := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, version)
	return version, err
}
