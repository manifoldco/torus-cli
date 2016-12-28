package registry

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
	result := &apitypes.Profile{}
	err := p.client.RoundTrip(ctx, "GET", "/profiles/"+name, nil, nil, &result)
	return result, err
}

// ListByID returns profiles looked up by User ID
func (p *ProfilesClient) ListByID(ctx context.Context, userIDs []identity.ID) ([]apitypes.Profile, error) {
	v := &url.Values{}
	for _, id := range userIDs {
		v.Add("id", id.String())
	}

	var results []apitypes.Profile
	err := p.client.RoundTrip(ctx, "GET", "/profiles", v, nil, &results)
	return results, err
}
