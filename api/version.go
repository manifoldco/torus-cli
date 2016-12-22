package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
)

type upstreamVersionClient struct {
	client RoundTripper
}

// VersionClient provides access to the daemon's /v1/version endpoint, for
// inspecting the daemon's release version.
type VersionClient struct {
	upstreamVersionClient
	client *apiRoundTripper
}

func newVersionClient(rt *apiRoundTripper) *VersionClient {
	return &VersionClient{upstreamVersionClient{rt}, rt}
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

// Get returns the registry's release version.
func (v *upstreamVersionClient) Get(ctx context.Context) (*apitypes.Version, error) {
	req, err := v.client.NewRequest("GET", "/version", nil, nil)
	if err != nil {
		return nil, err
	}

	version := &apitypes.Version{}
	_, err = v.client.Do(ctx, req, version)
	return version, err
}
