package api

import (
	"context"

	"github.com/arigatomachine/cli/apitypes"
)

// VersionClient provides access to the daemon's /v1/version endpoint, for
// inspecting the daemon's release version.
type VersionClient struct {
	client *Client
}

// Get returns the daemon's release version.
func (v *VersionClient) Get(ctx context.Context) (*apitypes.Version, *apitypes.Version, error) {
	req, _, err := v.client.NewRequest("GET", "/version", nil, nil, true)
	if err != nil {
		return nil, nil, err
	}

	registryVersion := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, registryVersion, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	req, _, err = v.client.NewRequest("GET", "/version", nil, nil, false)
	if err != nil {
		return nil, nil, err
	}

	daemonVersion := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, daemonVersion, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	return daemonVersion, registryVersion, nil
}
