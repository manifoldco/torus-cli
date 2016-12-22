package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// TeamsClient makes proxied requests to the registry's teams endpoints
type TeamsClient struct {
	client RoundTripper
}

// List retrieves all teams for an org based on the filtered values
func (t *TeamsClient) List(ctx context.Context, orgID *identity.ID, name string, teamType primitive.TeamType) ([]envelope.Team, error) {
	v := &url.Values{}

	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if name != "" {
		v.Set("name", name)
	}
	if teamType != primitive.AnyTeamType {
		v.Set("type", string(teamType))
	}

	req, err := t.client.NewRequest("GET", "/teams", v, nil)
	if err != nil {
		return nil, err
	}

	teams := []envelope.Team{}
	_, err = t.client.Do(ctx, req, &teams)
	return teams, err
}

// GetByOrg retrieves all teams for an org id
func (t *TeamsClient) GetByOrg(ctx context.Context, orgID *identity.ID) ([]envelope.Team, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	req, err := t.client.NewRequest("GET", "/teams", v, nil)
	if err != nil {
		return nil, err
	}

	teams := []envelope.Team{}
	_, err = t.client.Do(ctx, req, &teams)
	return teams, err
}

// GetByName retrieves the team with the specified name
func (t *TeamsClient) GetByName(ctx context.Context, orgID *identity.ID, name string) ([]envelope.Team, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())
	v.Set("name", name)

	req, err := t.client.NewRequest("GET", "/teams", v, nil)
	if err != nil {
		return nil, err
	}

	teams := []envelope.Team{}
	_, err = t.client.Do(ctx, req, &teams)
	return teams, err
}

// Create performs a request to create a new team object
func (t *TeamsClient) Create(ctx context.Context, orgID *identity.ID, name string,
	teamType primitive.TeamType) (*envelope.Team, error) {
	if orgID == nil {
		return nil, errors.New("invalid org")
	}

	teamBody := primitive.Team{
		Name:     name,
		OrgID:    orgID,
		TeamType: teamType,
	}

	ID, err := identity.NewMutable(&teamBody)
	if err != nil {
		return nil, err
	}

	team := envelope.Team{
		ID:      &ID,
		Version: 1,
		Body:    &teamBody,
	}

	req, err := t.client.NewRequest("POST", "/teams", nil, team)
	if err != nil {
		return nil, err
	}

	teamResult := &envelope.Team{}
	_, err = t.client.Do(ctx, req, teamResult)
	return teamResult, err
}
