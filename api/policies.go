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

// List retrieves relevant polciies by orgID and/or projectID
func (o *PoliciesClient) List(ctx context.Context, orgID, projectID *identity.ID) ([]PoliciesResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}
	if projectID != nil && projectID.Type() != 0 {
		v.Set("project_id", projectID.String())
	}

	req, _, err := o.client.NewRequest("GET", "/policies", v, nil, true)
	if err != nil {
		return nil, err
	}

	policies := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &policies, nil, nil)
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

// AttachmentsList retrieves all policy attachments for an org
func (o *PoliciesClient) AttachmentsList(ctx context.Context, orgID *identity.ID) ([]PolicyAttachmentResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, _, err := o.client.NewRequest("GET", "/policy-attachments", v, nil, true)
	if err != nil {
		return nil, err
	}

	attachments := make([]envelope.Unsigned, 1)
	_, err = o.client.Do(ctx, req, &attachments, nil, nil)
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
