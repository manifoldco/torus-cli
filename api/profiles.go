package api

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
)

// ProfilesClient makes proxied requests to the registry's profiles endpoints
type ProfilesClient struct {
	client RoundTripper
}

// ListByName returns profiles looked up by username
func (p *ProfilesClient) ListByName(ctx context.Context, name string) (*apitypes.Profile, error) {
	req, err := p.client.NewRequest("GET", "/profiles/"+name, nil, nil)
	if err != nil {
		return nil, err
	}

	result := &apitypes.Profile{}
	_, err = p.client.Do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListByID returns profiles looked up by User ID
func (p *ProfilesClient) ListByID(ctx context.Context, userIDs []identity.ID) (*[]apitypes.Profile, error) {
	v := &url.Values{}
	for _, id := range userIDs {
		v.Add("id", id.String())
	}

	req, err := p.client.NewRequest("GET", "/profiles", v, nil)
	if err != nil {
		return nil, err
	}

	results := []apitypes.Profile{}
	_, err = p.client.Do(ctx, req, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
