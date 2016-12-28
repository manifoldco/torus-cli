package registry

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
	client RoundTripper
}

type orgCreateRequest struct {
	Body struct {
		Name string `json:"name"`
	} `json:"body"`
}

// OrgTreeSegment is the payload returned for an org tree
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

	res := envelope.Org{}
	err := o.client.RoundTrip(ctx, "POST", "/orgs", nil, &org, &res)
	return &res, err
}

// Get returns the organization with the given ID.
func (o *OrgsClient) Get(ctx context.Context, orgID *identity.ID) (*envelope.Org, error) {
	org := envelope.Org{}
	err := o.client.RoundTrip(ctx, "GET", "/orgs/"+orgID.String(), nil, nil, &org)
	return &org, err
}

// GetByName retrieves an org by its name
func (o *OrgsClient) GetByName(ctx context.Context, name string) (*envelope.Org, error) {
	v := &url.Values{}
	if name == "" {
		return nil, errors.New("invalid org name")
	}
	v.Set("name", name)

	var orgs []envelope.Org
	err := o.client.RoundTrip(ctx, "GET", "/orgs", v, nil, &orgs)
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
	var orgs []envelope.Org
	err := o.client.RoundTrip(ctx, "GET", "/orgs", nil, nil, &orgs)
	return orgs, err
}

// RemoveMember removes a user from an org
func (o *OrgsClient) RemoveMember(ctx context.Context, orgID identity.ID,
	userID identity.ID) error {

	path := "/orgs/" + orgID.String() + "/members/" + userID.String()
	return o.client.RoundTrip(ctx, "DELETE", path, nil, nil, nil)
}

// GetTree returns an org tree
func (o *OrgsClient) GetTree(ctx context.Context, orgID identity.ID) ([]OrgTreeSegment, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	var segments []OrgTreeSegment
	err := o.client.RoundTrip(ctx, "GET", "/orgtree", v, nil, &segments)
	return segments, err
}
