package api

import (
	"context"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// OrgsClient makes proxied requests to the registry's orgs endpoints
type OrgsClient struct {
	client *Client
}

type orgCreateRequest struct {
	Body struct {
		Name string `json:"name"`
	} `json:"body"`
}

// OrgTreeSegment is the payload returns for an org tree
type OrgTreeSegment struct {
	Org      *primitive.Org      `json:"org"`
	Policies []primitive.Policy  `json:"policies"`
	Profiles []*apitypes.Profile `json:"profiles"`
	Teams    []*struct {
		Team              *envelope.Team               `json:"team"`
		Memberships       *[]envelope.Membership       `json:"memberships"`
		PolicyAttachments *[]envelope.PolicyAttachment `json:"policy_attachments"`
	} `json:"teams"`
}

// Create creates a new org with the given name. It returns the newly-created org.
func (o *OrgsClient) Create(ctx context.Context, name string) (*envelope.Org, error) {
	org := orgCreateRequest{}
	org.Body.Name = name

	req, _, err := o.client.NewRequest("POST", "/orgs", nil, &org, true)
	if err != nil {
		return nil, err
	}

	res := envelope.Org{}
	_, err = o.client.Do(ctx, req, &res, nil, nil)
	return &res, err
}

// GetByName retrieves an org by its named
func (o *OrgsClient) GetByName(ctx context.Context, name string) (*envelope.Org, error) {
	v := &url.Values{}
	if name == "" {
		return nil, errors.New("invalid org name")
	}
	v.Set("name", name)

	req, _, err := o.client.NewRequest("GET", "/orgs", v, nil, true)
	if err != nil {
		return nil, err
	}

	var orgs []envelope.Org
	_, err = o.client.Do(ctx, req, &orgs, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(orgs) < 1 {
		return nil, nil
	}

	return &orgs[0], nil
}

// List returns all organizations that the signed-in user has access to
func (o *OrgsClient) List(ctx context.Context) ([]envelope.Org, error) {
	req, _, err := o.client.NewRequest("GET", "/orgs", nil, nil, true)
	if err != nil {
		return nil, err
	}

	var orgs []envelope.Org
	_, err = o.client.Do(ctx, req, &orgs, nil, nil)
	return orgs, err
}

// RemoveMember removes a user from an org
func (o *OrgsClient) RemoveMember(ctx context.Context, orgID identity.ID,
	userID identity.ID) error {

	req, _, err := o.client.NewRequest("DELETE",
		"/orgs/"+orgID.String()+"/members/"+userID.String(), nil, nil, true)
	if err != nil {
		return err
	}

	_, err = o.client.Do(ctx, req, nil, nil, nil)
	return err
}

// GetTree returns an org tree
func (o *OrgsClient) GetTree(ctx context.Context, orgID identity.ID) ([]OrgTreeSegment, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())
	req, _, err := o.client.NewRequest("GET", "/orgtree", v, nil, true)
	if err != nil {
		return nil, err
	}

	segments := []OrgTreeSegment{}
	_, err = o.client.Do(ctx, req, &segments, nil, nil)
	return segments, err
}
