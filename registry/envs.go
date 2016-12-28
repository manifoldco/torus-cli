package registry

import (
	"context"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// EnvironmentsClient makes proxied requests to the registry's envs endpoints
type EnvironmentsClient struct {
	client RoundTripper
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

	env := envelope.Environment{
		ID:      &ID,
		Version: 1,
		Body:    &envBody,
	}

	return e.client.RoundTrip(ctx, "POST", "/envs", nil, env, nil)
}

// List retrieves relevant envs by name and/or orgID and/or projectID
func (e *EnvironmentsClient) List(ctx context.Context, orgIDs, projectIDs *[]*identity.ID, names *[]string) ([]envelope.Environment, error) {
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

	var envs []envelope.Environment
	err := e.client.RoundTrip(ctx, "GET", "/envs", v, nil, &envs)
	return envs, err
}
