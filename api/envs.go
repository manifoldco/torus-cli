package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// EnvironmentsClient makes proxied requests to the registry's envs endpoints
type EnvironmentsClient struct {
	client *Client
}

// EnvironmentResult is the payload returned for a env object
type EnvironmentResult struct {
	ID      *identity.ID           `json:"id"`
	Version uint8                  `json:"version"`
	Body    *primitive.Environment `json:"body"`
}

// Create generates a new env object for an org/project ID
func (e *EnvironmentsClient) Create(ctx context.Context, orgID, projectID *identity.ID, name string) error {
	if orgID == nil || projectID == nil {
		return errors.New("invalid org or project")
	}

	envBody := primitive.Environment{
		Name:      name,
		OrgID:     orgID,
		ProjectID: projectID,
	}

	ID, err := identity.NewMutable(&envBody)
	if err != nil {
		return err
	}

	env := apitypes.Environment{
		ID:      ID.String(),
		Version: 1,
		Body:    &envBody,
	}

	req, _, err := e.client.NewRequest("POST", "/envs", nil, env, true)
	if err != nil {
		return err
	}

	_, err = e.client.Do(ctx, req, nil, nil, nil)
	return err
}

// List retrieves relevant envs by name and/or orgID and/or projectID
func (e *EnvironmentsClient) List(ctx context.Context, orgIDs, projectIDs *[]*identity.ID, names *[]string) ([]EnvironmentResult, error) {
	v := &url.Values{}
	if orgIDs != nil {
		for _, id := range *orgIDs {
			v.Add("org_id", id.String())
		}
	}
	if projectIDs != nil {
		for _, id := range *projectIDs {
			v.Add("project_id", id.String())
		}
	}
	if names != nil {
		for _, n := range *names {
			v.Add("name", n)
		}
	}

	req, _, err := e.client.NewRequest("GET", "/envs", v, nil, true)
	if err != nil {
		return nil, err
	}

	envs := []EnvironmentResult{}
	_, err = e.client.Do(ctx, req, &envs, nil, nil)
	return envs, err
}
