package registry

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// PoliciesClient makes proxied requests to the registry's policies endpoints
type PoliciesClient struct {
	client RoundTripper
}

// Create creates a new policy
func (p *PoliciesClient) Create(ctx context.Context, policy *primitive.Policy) (*envelope.Policy, error) {

	ID, err := identity.NewMutable(policy)
	if err != nil {
		return nil, err
	}

	env := envelope.Policy{
		ID:      &ID,
		Version: 1,
		Body:    policy,
	}

	req, err := p.client.NewRequest("POST", "/policies", nil, env)
	if err != nil {
		return nil, err
	}

	res := envelope.Policy{}
	_, err = p.client.Do(ctx, req, &res)
	return &res, err
}

// List retrieves relevant policiies by orgID and/or name
func (p *PoliciesClient) List(ctx context.Context, orgID *identity.ID, name string) ([]envelope.Policy, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if name != "" {
		v.Set("name", name)
	}

	req, err := p.client.NewRequest("GET", "/policies", v, nil)
	if err != nil {
		return nil, err
	}

	policies := []envelope.Policy{}
	_, err = p.client.Do(ctx, req, &policies)
	return policies, err
}

// Attach attaches a policy to a team
func (p *PoliciesClient) Attach(ctx context.Context, org, policy, team *identity.ID) error {
	attachment := primitive.PolicyAttachment{
		OrgID:    org,
		OwnerID:  team,
		PolicyID: policy,
	}

	ID, err := identity.NewMutable(&attachment)
	if err != nil {
		return err
	}

	env := envelope.PolicyAttachment{
		ID:      &ID,
		Version: 1,
		Body:    &attachment,
	}

	req, err := p.client.NewRequest("POST", "/policy-attachments", nil, &env)
	if err != nil {
		return err
	}
	_, err = p.client.Do(ctx, req, nil)
	return err
}

// Detach deletes a specific attachment
func (p *PoliciesClient) Detach(ctx context.Context, attachmentID *identity.ID) error {
	req, err := p.client.NewRequest("DELETE", "/policy-attachments/"+attachmentID.String(), nil, nil)
	if err != nil {
		return err
	}
	_, err = p.client.Do(ctx, req, nil)
	return err
}

// AttachmentsList retrieves all policy attachments for an org
func (p *PoliciesClient) AttachmentsList(ctx context.Context, orgID, ownerID, policyID *identity.ID) ([]envelope.PolicyAttachment, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if ownerID != nil {
		v.Set("owner_id", ownerID.String())
	}
	if policyID != nil {
		v.Set("policy_id", policyID.String())
	}

	req, err := p.client.NewRequest("GET", "/policy-attachments", v, nil)
	if err != nil {
		return nil, err
	}

	attachments := []envelope.PolicyAttachment{}
	_, err = p.client.Do(ctx, req, &attachments)
	return attachments, err
}
