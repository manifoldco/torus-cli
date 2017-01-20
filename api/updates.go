package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
)

// UpdatesClient checks if there are updates available coming from
// the daemon update checker component.
type UpdatesClient struct {
	client *apiRoundTripper
}

// Check returns the latest updates check result, useful for detecting whether
// a newer version of Torus is available for download.
func (c *UpdatesClient) Check(ctx context.Context) (*apitypes.UpdateInfo, error) {
	var needsUpdate apitypes.UpdateInfo
	err := c.client.DaemonRoundTrip(ctx, "GET", "/updates", nil, nil, &needsUpdate, nil)
	return &needsUpdate, err
}
