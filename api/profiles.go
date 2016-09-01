package api

import (
	"context"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
)

// ProfilesClient makes proxied requests to the registry's profiles endpoints
type ProfilesClient struct {
	client *Client
}

// ListByID returns profiles looked up by User ID
func (o *ProfilesClient) ListByID(ctx context.Context, userIDs []identity.ID) ([]apitypes.Profile, error) {
	v := &url.Values{}
	for _, id := range userIDs {
		v.Add("id", id.String())
	}

	req, err := o.client.NewRequest("GET", "/profiles", v, nil, true)
	if err != nil {
		return nil, err
	}

	results := []apitypes.Profile{}
	_, err = o.client.Do(ctx, req, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
