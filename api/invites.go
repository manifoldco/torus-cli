package api

import (
	"context"
	"net/url"
	"time"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// InvitesClient makes proxied requests to the registry's invites endpoints
type InvitesClient struct {
	client *Client
}

// List all invites for a given org
func (i *InvitesClient) List(ctx context.Context, orgID *identity.ID, states []string) ([]envelope.OrgInvite, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	for _, state := range states {
		v.Add("state", state)
	}

	req, _, err := i.client.NewRequest("GET", "/org-invites", v, nil, true)
	if err != nil {
		return nil, err
	}

	invites := []envelope.OrgInvite{}
	_, err = i.client.Do(ctx, req, &invites, nil, nil)
	return invites, err
}

// Send creates a new org invitation
func (i *InvitesClient) Send(ctx context.Context, email string, orgID, inviterID identity.ID, teamIDs []identity.ID) error {
	now := time.Now()

	inviteBody := primitive.OrgInvite{
		OrgID:        &orgID,
		InviterID:    &inviterID,
		PendingTeams: teamIDs,
		Email:        email,
		Created:      &now,
		// Null values below
		InviteeID:  nil,
		ApproverID: nil,
		Accepted:   nil,
		Approved:   nil,
	}

	ID, err := identity.NewMutable(&inviteBody)
	if err != nil {
		return err
	}

	invite := envelope.OrgInvite{
		ID:      &ID,
		Version: 1,
		Body:    &inviteBody,
	}

	req, _, err := i.client.NewRequest("POST", "/org-invites", nil, &invite, true)
	if err != nil {
		return err
	}

	_, err = i.client.Do(ctx, req, nil, nil, nil)
	return err
}

// Accept executes the accept invite request
func (i *InvitesClient) Accept(ctx context.Context, org, email, code string) error {
	data := apitypes.InviteAccept{
		Org:   org,
		Email: email,
		Code:  code,
	}

	req, reqID, err := i.client.NewRequest("POST", "/org-invites/accept", nil, data, true)
	if err != nil {
		return err
	}

	_, err = i.client.Do(ctx, req, nil, &reqID, nil)
	return err
}

// Associate executes the associate invite request
func (i *InvitesClient) Associate(ctx context.Context, org, email, code string) (*envelope.OrgInvite, error) {
	// Same payload as accept, re-use type
	data := apitypes.InviteAccept{
		Org:   org,
		Email: email,
		Code:  code,
	}

	req, reqID, err := i.client.NewRequest("POST", "/org-invites/associate", nil, data, true)
	if err != nil {
		return nil, err
	}

	invite := envelope.OrgInvite{}
	_, err = i.client.Do(ctx, req, &invite, &reqID, nil)
	return &invite, err
}

// Approve executes the approve invite request
func (i *InvitesClient) Approve(ctx context.Context, inviteID identity.ID, output *ProgressFunc) error {
	req, reqID, err := i.client.NewRequest("POST", "/org-invites/"+inviteID.String()+"/approve", nil, nil, false)
	if err != nil {
		return err
	}

	_, err = i.client.Do(ctx, req, nil, &reqID, output)
	return err
}
