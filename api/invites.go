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
	for i, t := range invites {
		invites := InviteResult{}
		invites.ID = t.ID
		invites.Version = t.Version

		invitesBody, ok := t.Body.(*primitive.OrgInvite)
		if !ok {
			return nil, errors.New("invalid org invite body")
		}
		invites.Body = invitesBody
		invitesResults[i] = invites
	}

	return invitesResults, nil
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
	if err != nil {
		return err
	}

	return nil
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
