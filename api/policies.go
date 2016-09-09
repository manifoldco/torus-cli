package api

import (
	"context"
	"errors"
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

// List retrieves relevant policiies by orgID and/or projectID
func (p *PoliciesClient) List(ctx context.Context, orgID, projectID *identity.ID) ([]PoliciesResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if projectID != nil && projectID.Type() != 0 {
		v.Set("project_id", projectID.String())
	}

	req, _, err := p.client.NewRequest("GET", "/policies", v, nil, true)
	if err != nil {
		return nil, err
	}

	policies := make([]envelope.Unsigned, 1)
	_, err = p.client.Do(ctx, req, &policies, nil, nil)
	if err != nil {
		return nil, err
	}

	policyResults := make([]PoliciesResult, len(policies))
	for i, t := range policies {
		policy := PoliciesResult{}
		policy.ID = t.ID
		policy.Version = t.Version

		policyBody, ok := t.Body.(*primitive.Policy)
		if !ok {
			return nil, errors.New("invalid policy body")
		}
		policy.Body = policyBody
		policyResults[i] = policy
	}

	return policyResults, nil
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
	if err != nil {
		return err
	}

	return nil
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

	var attachments []envelope.Unsigned
	_, err = p.client.Do(ctx, req, &attachments, nil, nil)
	if err != nil {
		return nil, err
	}

	policyAttachmentResults := make([]PolicyAttachmentResult, len(attachments))
	for i, a := range attachments {
		attachment := PolicyAttachmentResult{}
		attachment.ID = a.ID
		attachment.Version = a.Version

		policyBody, ok := a.Body.(*primitive.PolicyAttachment)
		if !ok {
			return nil, errors.New("invalid policy attachment body")
		}
		attachment.Body = policyBody
		policyAttachmentResults[i] = attachment
	}

	return policyAttachmentResults, nil
}
