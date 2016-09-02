package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// ProjectsClient makes proxied requests to the registry's projects endpoints
type ProjectsClient struct {
	client *Client
}

// ProjectResult is the payload returned for a project object
type ProjectResult struct {
	ID      *identity.ID       `json:"id"`
	Version uint8              `json:"version"`
	Body    *primitive.Project `json:"body"`
}

// List retrieves relevant projects by name and/or orgID
func (o *ProjectsClient) List(ctx context.Context, orgID *identity.ID, name *string) ([]ProjectResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if name != nil {
		v.Set("name", *name)
	}

	req, err := o.client.NewRequest("GET", "/projects", v, nil, true)
	if err != nil {
		return nil, err
	}

	projects := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &projects)
	if err != nil {
		return nil, err
	}

	projectResults := make([]ProjectResult, len(projects))
	for i, t := range projects {
		project := ProjectResult{}
		project.ID = t.ID
		project.Version = t.Version

		projectBody, ok := t.Body.(*primitive.Project)
		if !ok {
			return nil, errors.New("invalid project body")
		}
		project.Body = projectBody
		projectResults[i] = project
	}

	return projectResults, nil
}
