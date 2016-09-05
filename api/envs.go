package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/envelope"
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

	ID, err := identity.Mutable(&envBody)
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
	if err != nil {
		return err
	}

	return nil
}

// List retrieves relevant envs by name and/or orgID and/or projectID
func (e *EnvironmentsClient) List(ctx context.Context, orgID, projectID *identity.ID, name *string) ([]EnvironmentResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if projectID != nil && projectID.Type() != 0 {
		v.Set("project_id", projectID.String())
	}
	if name != nil {
		v.Set("name", *name)
	}

	req, _, err := e.client.NewRequest("GET", "/envs", v, nil, true)
	if err != nil {
		return nil, err
	}

	envs := make([]envelope.Unsigned, 1)
	_, err = e.client.Do(ctx, req, &envs, nil, nil)
	if err != nil {
		return nil, err
	}

	envResults := make([]EnvironmentResult, len(envs))
	for i, t := range envs {
		env := EnvironmentResult{}
		env.ID = t.ID
		env.Version = t.Version

		envBody, ok := t.Body.(*primitive.Environment)
		if !ok {
			return nil, errors.New("invalid env body")
		}
		env.Body = envBody
		envResults[i] = env
	}

	return envResults, nil
}
