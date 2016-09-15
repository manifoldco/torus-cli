package api

import (
	"context"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// PoliciesClient makes proxied requests to the registry's policies endpoints
type PoliciesClient struct {
	client *Client
}

// PolicyAttachmentResult is the payload returned for a policy_attachment object
type PolicyAttachmentResult struct {
	ID      *identity.ID                `json:"id"`
	Version uint8                       `json:"version"`
	Body    *primitive.PolicyAttachment `json:"body"`
}

// PoliciesResult is the payload returned for a policy object
type PoliciesResult struct {
	ID      *identity.ID      `json:"id"`
	Version uint8             `json:"version"`
	Body    *primitive.Policy `json:"body"`
}

// Create creates a new policy
func (p *PoliciesClient) Create(ctx context.Context, policy *primitive.Policy) (*PoliciesResult, error) {

	ID, err := identity.Mutable(policy)
	if err != nil {
		return nil, err
	}

	env := envelope.Unsigned{
		ID:      &ID,
		Version: 1,
		Body:    policy,
	}

	req, _, err := p.client.NewRequest("POST", "/policies", nil, env, true)
	if err != nil {
		return nil, err
	}

	res := PoliciesResult{}
	_, err = p.client.Do(ctx, req, &res, nil, nil)
	return &res, err
}

// List retrieves relevant policiies by orgID and/or name
func (p *PoliciesClient) List(ctx context.Context, orgID *identity.ID, name string) ([]PoliciesResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if name != "" {
		v.Set("name", name)
	}

	req, _, err := p.client.NewRequest("GET", "/policies", v, nil, true)
	if err != nil {
		return nil, err
	}

	policies := []PoliciesResult{}
	_, err = p.client.Do(ctx, req, &policies, nil, nil)
	return policies, err
}

// Attach attaches a policy to a team
func (p *PoliciesClient) Attach(ctx context.Context, org, policy, team *identity.ID) error {
	attachment := primitive.PolicyAttachment{
		OrgID:    org,
		OwnerID:  team,
		PolicyID: policy,
	}

	ID, err := identity.Mutable(&attachment)
	if err != nil {
		return err
	}

	env := envelope.Unsigned{
		ID:      &ID,
		Version: 1,
		Body:    &attachment,
	}

	req, _, err := p.client.NewRequest("POST", "/policy-attachments", nil, &env, true)
	if err != nil {
		return err
	}
	_, err = p.client.Do(ctx, req, nil, nil, nil)
	return err
}

// Detach deletes a specific attachment
func (p *PoliciesClient) Detach(ctx context.Context, attachmentID *identity.ID) error {
	req, _, err := p.client.NewRequest("DELETE", "/policy-attachments/"+attachmentID.String(), nil, nil, true)
	if err != nil {
		return err
	}
	_, err = p.client.Do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// AttachmentsList retrieves all policy attachments for an org
func (p *PoliciesClient) AttachmentsList(ctx context.Context, orgID, ownerID, policyID *identity.ID) ([]PolicyAttachmentResult, error) {
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

	req, _, err := p.client.NewRequest("GET", "/policy-attachments", v, nil, true)
	if err != nil {
		return nil, err
	}

	attachments := []PolicyAttachmentResult{}
	_, err = p.client.Do(ctx, req, &attachments, nil, nil)
	return attachments, err
}
