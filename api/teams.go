package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// TeamsClient makes proxied requests to the registry's teams endpoints
type TeamsClient struct {
	client *Client
}

// TeamResult is the payload returned for a team object
type TeamResult struct {
	ID      *identity.ID    `json:"id"`
	Version uint8           `json:"version"`
	Body    *primitive.Team `json:"body"`
}

// GetByOrg retrieves all teams for an org id
func (o *TeamsClient) GetByOrg(ctx context.Context, orgID *identity.ID) ([]TeamResult, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	req, _, err := o.client.NewRequest("GET", "/teams", v, nil, true)
	if err != nil {
		return nil, err
	}

	teams := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &teams, nil, nil)
	if err != nil {
		return nil, err
	}

	return buildTeamResults(teams)
}

// GetByName retrieves the team with the specified name
func (o *TeamsClient) GetByName(ctx context.Context, orgID *identity.ID, name string) ([]TeamResult, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())
	v.Set("name", name)

	req, _, err := o.client.NewRequest("GET", "/teams", v, nil, true)
	if err != nil {
		return nil, err
	}

	teams := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &teams, nil, nil)
	if err != nil {
		return nil, err
	}

	return buildTeamResults(teams)
}

func buildTeamResults(teams []envelope.Unsigned) ([]TeamResult, error) {
	teamResults := make([]TeamResult, len(teams))
	for i, t := range teams {
		team := TeamResult{}
		team.ID = t.ID
		team.Version = t.Version

		teamBody, ok := t.Body.(*primitive.Team)
		if !ok {
			return nil, errors.New("invalid team body")
		}
		team.Body = teamBody
		teamResults[i] = team
	}

	return teamResults, nil
}
