package registry

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// ProjectsClient represents the `/projects` registry endpoint, for
// manipulating projects.
type ProjectsClient struct {
	client *Client
}

// List returns a list of all Projects within the given org.
func (p *ProjectsClient) List(ctx context.Context, orgID *identity.ID) ([]envelope.Project, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, err := p.client.NewRequest("GET", "/projects", v, nil)
	if err != nil {
		return nil, err
	}

	var projects []envelope.Project
	_, err = p.client.Do(ctx, req, &projects)
	return projects, err
}
