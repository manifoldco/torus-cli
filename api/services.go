package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
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
func (o *ServicesClient) List(ctx context.Context, orgID, projectID *identity.ID, name *string) ([]ServiceResult, error) {
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

	req, _, err := o.client.NewRequest("GET", "/services", v, nil, true)
	if err != nil {
		return nil, err
	}

	services := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &services, nil, nil)
	if err != nil {
		return nil, err
	}

	serviceResults := make([]ServiceResult, len(services))
	for i, t := range services {
		service := ServiceResult{}
		service.ID = t.ID
		service.Version = t.Version

		serviceBody, ok := t.Body.(*primitive.Service)
		if !ok {
			return nil, errors.New("invalid service body")
		}
		service.Body = serviceBody
		serviceResults[i] = service
	}

	return serviceResults, nil
}
