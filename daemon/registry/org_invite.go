package registry

import (
	"context"
	"errors"
	"log"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// OrgInviteClient represents the `/org-invites` registry endpoint, used for
// sending, accepting, and approving invitations to organizations in Torus.
type OrgInviteClient struct {
	client *Client
}

// Approve sends an approval notification to the registry regarding a specific
// invitation.
func (o *OrgInviteClient) Approve(ctx context.Context, inviteID *identity.ID) (*envelope.OrgInvite, error) {

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
func (o *OrgInviteClient) Get(ctx context.Context, inviteID *identity.ID) (*envelope.OrgInvite, error) {
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
func (o *OrgInviteClient) List(ctx context.Context, orgID *identity.ID, states []string, email string) ([]envelope.OrgInvite, error) {
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
