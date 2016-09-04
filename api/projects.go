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

type projectCreateRequest struct {
	Body struct {
		OrgID *identity.ID `json:"org_id"`
		Name  string       `json:"name"`
	} `json:"body"`
}

// Create creates a new project with the given name within the given org
func (p *ProjectsClient) Create(ctx context.Context, org *identity.ID, name string) (*ProjectResult, error) {
	project := projectCreateRequest{}
	project.Body.OrgID = org
	project.Body.Name = name

	req, _, err := p.client.NewRequest("POST", "/projects", nil, &project, true)
	if err != nil {
		return nil, err
	}

	res := ProjectResult{}
	_, err = p.client.Do(ctx, req, &res, nil, nil)
	return &res, err
}

// List retrieves relevant projects by name and/or orgID
func (p *ProjectsClient) List(ctx context.Context, orgID *identity.ID, name *string) ([]ProjectResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if name != nil {
		v.Set("name", *name)
	}

	req, _, err := p.client.NewRequest("GET", "/projects", v, nil, true)
	if err != nil {
		return nil, err
	}

	projects := make([]envelope.Unsigned, 1)
	_, err = p.client.Do(ctx, req, &projects, nil, nil)
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
