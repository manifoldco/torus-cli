package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// MembershipsClient makes proxied requests to the registry's team memberships
// endpoints.
type MembershipsClient struct {
	client *Client
}

// MembershipResult is the payload returned for team membership association.
type MembershipResult struct {
	ID      *identity.ID
	Version uint8
	Body    *primitive.Membership
}

// List returns all team membership associations for the given user id within
// the given org id.
func (m *MembershipsClient) List(ctx context.Context, org, user *identity.ID) ([]MembershipResult, error) {
	v := &url.Values{}
	v.Set("org_id", org.String())
	v.Set("owner_id", user.String())

	req, err := m.client.NewRequest("GET", "/memberships", v, nil, true)
	if err != nil {
		return nil, err
	}

	envs := []envelope.Unsigned{}
	_, err = m.client.Do(ctx, req, &envs)
	if err != nil {
		return nil, err
	}

	memberships := make([]MembershipResult, len(envs))
	for i, env := range envs {
		membershipBody, ok := env.Body.(*primitive.Membership)
		if !ok {
			return nil, errors.New("invalid membership body")
		}
		memberships[i] = MembershipResult{
			ID:      env.ID,
			Version: env.Version,
			Body:    membershipBody,
		}
	}

	return memberships, err
}
