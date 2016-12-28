package registry

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// ProjectsClient makes proxied requests to the registry's projects endpoints
type ProjectsClient struct {
	client RoundTripper
}

type projectCreateRequest struct {
	Body struct {
		OrgID *identity.ID `json:"org_id"`
		Name  string       `json:"name"`
	} `json:"body"`
}

// ProjectTreeSegment is the payload returned for a project tree
type ProjectTreeSegment struct {
	Org      *envelope.Org           `json:"org"`
	Envs     []*envelope.Environment `json:"envs"`
	Services []*envelope.Service     `json:"services"`
	Projects []envelope.Project      `json:"projects"`
	Profiles []*apitypes.Profile     `json:"profiles"`
}

// Create creates a new project with the given name within the given org
func (p *ProjectsClient) Create(ctx context.Context, org *identity.ID, name string) (*envelope.Project, error) {
	project := projectCreateRequest{}
	project.Body.OrgID = org
	project.Body.Name = name

	res := envelope.Project{}
	err := p.client.RoundTrip(ctx, "POST", "/projects", nil, &project, &res)
	return &res, err
}

// Search retrieves relevant projects by name and/or orgID
func (p *ProjectsClient) Search(ctx context.Context, orgIDs *[]*identity.ID, names *[]string) ([]envelope.Project, error) {
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

	var projects []envelope.Project
	err := p.client.RoundTrip(ctx, "GET", "/projects", v, nil, &projects)
	return projects, err
}

// List returns a list of all Projects within the given org.
func (p *ProjectsClient) List(ctx context.Context, orgID *identity.ID) ([]envelope.Project, error) {
	orgs := []*identity.ID{orgID}
	return p.Search(ctx, &orgs, nil)
}

// GetTree returns a project tree
func (p *ProjectsClient) GetTree(ctx context.Context, orgID *identity.ID) ([]ProjectTreeSegment, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	var segments []ProjectTreeSegment
	err := p.client.RoundTrip(ctx, "GET", "/projecttree", v, nil, &segments)
	return segments, err
}
