package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// MembershipsClient makes proxied requests to the registry's team memberships
// endpoints.
type MembershipsClient struct {
	client *Client
}

// List returns all team membership associations for the given user id within
// the given org id.
func (m *MembershipsClient) List(ctx context.Context, org, user, team *identity.ID) ([]envelope.Membership, error) {
	v := &url.Values{}
	v.Set("org_id", org.String())
	if user != nil {
		v.Set("owner_id", user.String())
	}
	if team != nil {
		v.Set("team_id", team.String())
	}

	req, _, err := m.client.NewRequest("GET", "/memberships", v, nil, true)
	if err != nil {
		return nil, err
	}

	memberships := []envelope.Membership{}
	_, err = m.client.Do(ctx, req, &memberships, nil, nil)
	return memberships, err
}

// Create requests addition of a user to a team
func (m *MembershipsClient) Create(ctx context.Context, userID, orgID, teamID *identity.ID) error {
	if orgID == nil {
		return errors.New("invalid org")
	}
	if userID == nil {
		return errors.New("invalid user")
	}
	if teamID == nil {
		return errors.New("invalid team")
	}

	membershipBody := primitive.Membership{
		OwnerID: userID,
		OrgID:   orgID,
		TeamID:  teamID,
	}

	ID, err := identity.NewMutable(&membershipBody)
	if err != nil {
		return err
	}

	membership := envelope.Membership{
		ID:      &ID,
		Version: 1,
		Body:    &membershipBody,
	}

	req, _, err := m.client.NewRequest("POST", "/memberships", nil, membership, true)
	if err != nil {
		return err
	}

	_, err = m.client.Do(ctx, req, nil, nil, nil)
	return err
}

// Delete requests deletion of a specific membership row by ID
func (m *MembershipsClient) Delete(ctx context.Context, membership *identity.ID) error {
	req, _, err := m.client.NewRequest("DELETE", "/memberships/"+membership.String(), nil, nil, true)
	if err != nil {
		return err
	}

	_, err = m.client.Do(ctx, req, nil, nil, nil)
	return err
}
