package api

import (
	"context"
	"time"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// InvitesClient makes proxied requests to the registry's invites endpoints
type InvitesClient struct {
	client *Client
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

	req, err := i.client.NewRequest("POST", "/org-invites", nil, &invite, true)
	if err != nil {
		return err
	}

	_, err = i.client.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
