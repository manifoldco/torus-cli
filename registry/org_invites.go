package registry

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// OrgInvitesClient represents the `/org-invites` registry endpoint, used for
// sending, accepting, and approving invitations to organizations in Torus.
type OrgInvitesClient struct {
	client RoundTripper
}

// Approve sends an approval notification to the registry regarding a specific
// invitation.
func (o *OrgInvitesClient) Approve(ctx context.Context, inviteID *identity.ID) (*envelope.OrgInvite, error) {
	invite := envelope.OrgInvite{}
	path := "/org-invites/" + inviteID.String() + "/approve"
	err := o.client.RoundTrip(ctx, "POST", path, nil, nil, &invite)
	return &invite, err
}

// Get returns a specific Org Invite based on it's ID
func (o *OrgInvitesClient) Get(ctx context.Context, inviteID *identity.ID) (*envelope.OrgInvite, error) {
	if inviteID == nil {
		return nil, errors.New("an inviteID must be provided")
	}

	invite := envelope.OrgInvite{}
	path := "/org-invites/" + inviteID.String()
	err := o.client.RoundTrip(ctx, "GET", path, nil, nil, &invite)
	return &invite, err
}

// List lists all invites for a given org with the given states
func (o *OrgInvitesClient) List(ctx context.Context, orgID *identity.ID, states []string, email string) ([]envelope.OrgInvite, error) {
	v := &url.Values{}
	v.Set("org_id", orgID.String())

	for _, state := range states {
		v.Add("state", state)
	}

	if email != "" {
		v.Add("email", email)
	}

	var invites []envelope.OrgInvite
	err := o.client.RoundTrip(ctx, "GET", "/org-invites", v, nil, &invites)
	return invites, err
}

// Accept executes the accept invite request
func (o *OrgInvitesClient) Accept(ctx context.Context, org, email, code string) error {
	data := apitypes.InviteAccept{
		Org:   org,
		Email: email,
		Code:  code,
	}

	return o.client.RoundTrip(ctx, "POST", "/org-invites/accept", nil, &data, nil)
}

// Send creates a new org invitation
func (o *OrgInvitesClient) Send(ctx context.Context, email string, orgID, inviterID identity.ID, teamIDs []identity.ID) error {
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

	return o.client.RoundTrip(ctx, "POST", "/org-invites", nil, &invite, nil)
}

// Associate executes the associate invite request
func (o *OrgInvitesClient) Associate(ctx context.Context, org, email, code string) (*envelope.OrgInvite, error) {
	// Same payload as accept, re-use type
	data := apitypes.InviteAccept{
		Org:   org,
		Email: email,
		Code:  code,
	}

	invite := envelope.OrgInvite{}
	err := o.client.RoundTrip(ctx, "POST", "/org-invites/associate", nil, &data, &invite)
	return &invite, err
}
