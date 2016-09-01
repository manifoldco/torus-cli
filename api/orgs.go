package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// OrgsClient makes proxied requests to the registry's orgs endpoints
type OrgsClient struct {
	client *Client
}

// OrgResult is the payload returned for an org object
type OrgResult struct {
	ID      *identity.ID   `json:"id"`
	Version uint8          `json:"version"`
	Body    *primitive.Org `json:"body"`
}

// GetByName retrieves an org by its named
func (o *OrgsClient) GetByName(ctx context.Context, name string) (*OrgResult, error) {
	v := &url.Values{}
	v.Set("name", name)

	req, err := o.client.NewRequest("GET", "/orgs", v, nil, true)
	if err != nil {
		return nil, err
	}

	orgs := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &orgs)
	if err != nil {
		return nil, err
	}

	org := OrgResult{}
	org.ID = orgs[0].ID
	org.Version = orgs[0].Version

	orgBody, ok := orgs[0].Body.(*primitive.Org)
	if !ok {
		return nil, errors.New("invalid org body")
	}
	org.Body = orgBody

	return &org, nil
}
