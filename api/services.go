package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// ServicesClient makes proxied requests to the registry's services endpoints
type ServicesClient struct {
	client *Client
}

// ServiceResult is the payload returned for a service object
type ServiceResult struct {
	ID      *identity.ID       `json:"id"`
	Version uint8              `json:"version"`
	Body    *primitive.Service `json:"body"`
}

// List retrieves relevant services by name and/or orgID and/or projectID
func (s *ServicesClient) List(ctx context.Context, orgID, projectID *identity.ID, name *string) ([]ServiceResult, error) {
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

	req, _, err := s.client.NewRequest("GET", "/services", v, nil, true)
	if err != nil {
		return nil, err
	}

	services := []ServiceResult{}
	_, err = s.client.Do(ctx, req, &services, nil, nil)
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

	service := apitypes.Service{
		ID:      &ID,
		Version: 1,
		Body:    &serviceBody,
	}

	req, _, err := s.client.NewRequest("POST", "/services", nil, service, true)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil, nil, nil)
	return err
}
