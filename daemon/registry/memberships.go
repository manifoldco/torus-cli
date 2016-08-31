package registry

import (
	"context"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
)

// MembershipsClient represents the `/memberships` registry
// endpoint, used for accessing the relationship between users,
// organization, and teams.
type MembershipsClient struct {
	client *Client
}

// List returns all memberships for a given organization, team, or user/machine
func (m *MembershipsClient) List(ctx context.Context, orgID *identity.ID,
	teamID *identity.ID, ownerID *identity.ID) ([]envelope.Unsigned, error) {

	query := &url.Values{}
	if orgID != nil {
		query.Set("org_id", orgID.String())
	}
	if teamID != nil {
		query.Set("team_id", teamID.String())
	}
	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	req, err := m.client.NewRequest("GET", "/memberships", query, nil)
	if err != nil {
		log.Printf("could not build GET /memberships request: %s", err)
		return nil, err
	}

	memberships := []envelope.Unsigned{}
	_, err = m.client.Do(ctx, req, &memberships)
	if err != nil {
		log.Printf("could not perform GET /memberships: %s", err)
		return nil, err
	}

	return memberships, nil
}
