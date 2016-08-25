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
func (v *VersionClient) Get(ctx context.Context) (*apitypes.Version, error) {
	req, err := v.client.NewRequest("GET", "/version", nil, nil)
	if err != nil {
		return nil, err
	}

	resp := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
