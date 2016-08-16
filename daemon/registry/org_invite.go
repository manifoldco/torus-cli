package registry

import (
	"context"
	"errors"
	"log"

	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
)

// OrgInviteClient represents the `/org-invites` registry endpoint, used for
// sending, accepting, and approving invitations to organizations in Arigato.
type OrgInviteClient struct {
	client *Client
}

// Approve sends an approval notification to the registry regarding a specific
// invitation.
func (o *OrgInviteClient) Approve(inviteID *identity.ID) (*envelope.Unsigned, error) {

	path := "/org-invites/" + inviteID.String() + "/approve"
	req, err := o.client.NewRequest("POST", path, nil, nil)
	if err != nil {
		log.Printf(
			"Error building POST /org-invites/:id/approve api request: %s", err)
		return nil, err
	}

	invite := envelope.Unsigned{}
	_, err = o.client.Do(context.TODO(), req, &invite)
	if err != nil {
		log.Printf("Error performing POST /org-invites/:id/accept: %s", err)
		return nil, err
	}

	return &invite, nil
}

// Get returns a specific Org Invite based on it's ID
func (o *OrgInviteClient) Get(inviteID *identity.ID) (*envelope.Unsigned, error) {
	if inviteID == nil {
		return nil, errors.New("an inviteID must be provided")
	}

	path := "/org-invites/" + inviteID.String()
	req, err := o.client.NewRequest("GET", path, nil, nil)
	if err != nil {
		log.Printf("Error building GET /org-invites/:id request: %s", err)
		return nil, err
	}

	invite := envelope.Unsigned{}
	_, err = o.client.Do(context.TODO(), req, &invite)
	if err != nil {
		log.Printf("Error performing GET /org-invites/:id request: %s", err)
		return nil, err
	}

	return &invite, nil
}
