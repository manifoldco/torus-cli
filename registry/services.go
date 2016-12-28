package registry

import (
	"context"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// ServicesClient makes proxied requests to the registry's services endpoints
type ServicesClient struct {
	client RoundTripper
}

// List retrieves relevant services by name and/or orgID and/or projectID
func (s *ServicesClient) List(ctx context.Context, orgIDs, projectIDs *[]*identity.ID, names *[]string) ([]envelope.Service, error) {
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

	var services []envelope.Service
	err := s.client.RoundTrip(ctx, "GET", "/services", v, nil, &services)
	return services, err
}

// Create performs a request to create a new service object
func (s *ServicesClient) Create(ctx context.Context, orgID, projectID *identity.ID, name string) error {
	if orgID == nil || projectID == nil {
		return errors.New("invalid org or project")
	}

	serviceBody := primitive.Service{
		Name:      name,
		OrgID:     orgID,
		ProjectID: projectID,
	}

	ID, err := identity.NewMutable(&serviceBody)
	if err != nil {
		return err
	}

	service := envelope.Service{
		ID:      &ID,
		Version: 1,
		Body:    &serviceBody,
	}

	return s.client.RoundTrip(ctx, "POST", "/services", nil, service, nil)
}
