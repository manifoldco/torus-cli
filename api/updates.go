package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
)

// UpdatesClient checks if there are updates available coming from
// the daemon update checker component.
type UpdatesClient struct {
	client *Client
}

// Check returns the latest updates check result, useful for detecting whether
// a newer version of Torus is available for download.
func (c *UpdatesClient) Check(ctx context.Context) (*apitypes.Updates, error) {
	req, _, err := c.client.NewRequest("GET", "/updates", nil, nil, false)
	if err != nil {
		return nil, err
	}

	var needsUpdate apitypes.Updates
	if _, err := c.client.Do(ctx, req, &needsUpdate, nil, nil); err != nil {
		return nil, err
	}

	return &needsUpdate, nil
}
