package api

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// InvitesClient makes proxied requests to the registry's invites endpoints
type InvitesClient struct {
	client *Client
}

// InviteResult is the payload returned for a team object
type InviteResult struct {
	ID      *identity.ID         `json:"id"`
	Version uint8                `json:"version"`
	Body    *primitive.OrgInvite `json:"body"`
}

// List all invites for a given org
func (i *InvitesClient) List(ctx context.Context, orgID *identity.ID, states []string) ([]InviteResult, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	for _, state := range states {
		v.Add("state", state)
	}

	req, _, err := i.client.NewRequest("GET", "/org-invites", v, nil, true)
	if err != nil {
		return nil, err
	}

	var invites []envelope.Unsigned
	_, err = i.client.Do(ctx, req, &invites, nil, nil)
	if err != nil {
		return nil, err
	}

	invitesResults := make([]InviteResult, len(invites))
	for i, inv := range invites {
		invite, err := convertInviteResults(inv)
		if err != nil {
			return nil, err
		}
		invitesResults[i] = *invite
	}
	return invitesResults, nil
}

func convertInviteResults(i envelope.Unsigned) (*InviteResult, error) {
	invite := InviteResult{}
	invite.ID = i.ID
	invite.Version = i.Version

	inviteBody, ok := i.Body.(*primitive.OrgInvite)
	if !ok {
		return nil, errors.New("invalid org invite body")
	}
	invite.Body = inviteBody
	return &invite, nil
}

// Send creates a new org invitation
func (i *InvitesClient) Send(ctx context.Context, email string, orgID, inviterID identity.ID, teamIDs []identity.ID) error {
	invite := apitypes.OrgInvite{}
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

	ID, err := identity.Mutable(&inviteBody)
	if err != nil {
		return err
	}

	invite.ID = ID.String()
	invite.Version = 1
	invite.Body = &inviteBody

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
	if err != nil {
		return err
	}

	return nil
}

// Associate executes the associate invite request
func (i *InvitesClient) Associate(ctx context.Context, org, email, code string) (*InviteResult, error) {
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

	var invites envelope.Unsigned
	_, err = i.client.Do(ctx, req, &invites, &reqID, nil)
	if err != nil {
		return nil, err
	}

	result, err := convertInviteResults(invites)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("invalid org invite response")
	}

	return result, nil
}

// Approve executes the approve invite request
func (i *InvitesClient) Approve(ctx context.Context, inviteID identity.ID, output *ProgressFunc) error {
	req, reqID, err := i.client.NewRequest("POST", "/org-invites/"+inviteID.String()+"/approve", nil, nil, false)
	if err != nil {
		return err
	}

	_, err = i.client.Do(ctx, req, nil, &reqID, output)
	if err != nil {
		return err
	}

	return nil
}
