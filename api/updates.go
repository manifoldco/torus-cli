package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
)

type UpdatesClient struct {
	client *Client
}

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
