package api

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
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

// ProjectTreeSegment is the payload returned for a project tree
type ProjectTreeSegment struct {
	Org      *OrgResult              `json:"org"`
	Envs     []*apitypes.Environment `json:"envs"`
	Services []*apitypes.Service     `json:"services"`
	Projects []*ProjectResult        `json:"projects"`
	Profiles []*apitypes.Profile     `json:"profiles"`
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
func (p *ProjectsClient) List(ctx context.Context, orgIDs *[]*identity.ID, names *[]string) ([]ProjectResult, error) {
	v := &url.Values{}
	if orgIDs != nil {
		for _, id := range *orgIDs {
			v.Add("org_id", id.String())
		}
	}
	if names != nil {
		for _, n := range *names {
			v.Add("name", n)
		}
	}

	req, _, err := p.client.NewRequest("GET", "/projects", v, nil, true)
	if err != nil {
		return nil, err
	}

	projects := []ProjectResult{}
	_, err = p.client.Do(ctx, req, &projects, nil, nil)
	return projects, err
}

// GetTree returns a project tree
func (p *ProjectsClient) GetTree(ctx context.Context, orgID *identity.ID) ([]ProjectTreeSegment, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())
	req, _, err := p.client.NewRequest("GET", "/projecttree", v, nil, true)
	if err != nil {
		return nil, err
	}

	segments := []ProjectTreeSegment{}
	_, err = p.client.Do(ctx, req, &segments, nil, nil)
	return segments, err
}
