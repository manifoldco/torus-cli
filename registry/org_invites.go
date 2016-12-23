package registry

import (
	"context"
	"errors"
	"log"
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

	path := "/org-invites/" + inviteID.String() + "/approve"
	req, err := o.client.NewRequest("POST", path, nil, nil)
	if err != nil {
		log.Printf(
			"Error building POST /org-invites/:id/approve api request: %s", err)
		return nil, err
	}

	invite := envelope.OrgInvite{}
	_, err = o.client.Do(ctx, req, &invite)
	if err != nil {
		log.Printf("Error performing POST /org-invites/:id/accept: %s", err)
		return nil, err
	}

	return &invite, nil
}

// Get returns a specific Org Invite based on it's ID
func (o *OrgInvitesClient) Get(ctx context.Context, inviteID *identity.ID) (*envelope.OrgInvite, error) {
	if inviteID == nil {
		return nil, errors.New("an inviteID must be provided")
	}

	path := "/org-invites/" + inviteID.String()
	req, err := o.client.NewRequest("GET", path, nil, nil)
	if err != nil {
		log.Printf("Error building GET /org-invites/:id request: %s", err)
		return nil, err
	}

	invite := envelope.OrgInvite{}
	_, err = o.client.Do(ctx, req, &invite)
	if err != nil {
		log.Printf("Error performing GET /org-invites/:id request: %s", err)
		return nil, err
	}

	return &invite, nil
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

	req, err := o.client.NewRequest("GET", "/org-invites", v, nil)
	if err != nil {
		return nil, err
	}

	var invites []envelope.OrgInvite
	_, err = o.client.Do(ctx, req, &invites)
	return invites, err
}

// Accept executes the accept invite request
func (o *OrgInvitesClient) Accept(ctx context.Context, org, email, code string) error {
	data := apitypes.InviteAccept{
		Org:   org,
		Email: email,
		Code:  code,
	}

	req, err := o.client.NewRequest("POST", "/org-invites/accept", nil, data)
	if err != nil {
		return err
	}

	_, err = o.client.Do(ctx, req, nil)
	return err
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

	req, err := o.client.NewRequest("POST", "/org-invites", nil, &invite)
	if err != nil {
		return err
	}

	_, err = o.client.Do(ctx, req, nil)
	return err
}

// Associate executes the associate invite request
func (o *OrgInvitesClient) Associate(ctx context.Context, org, email, code string) (*envelope.OrgInvite, error) {
	// Same payload as accept, re-use type
	data := apitypes.InviteAccept{
		Org:   org,
		Email: email,
		Code:  code,
	}

	req, err := o.client.NewRequest("POST", "/org-invites/associate", nil, data)
	if err != nil {
		return nil, err
	}

	invite := envelope.OrgInvite{}
	_, err = o.client.Do(ctx, req, &invite)
	return &invite, err
}
