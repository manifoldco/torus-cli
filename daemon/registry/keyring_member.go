package registry

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/envelope"
)

// KeyringMemberClientV1 represents the `/keyring-members` registry endpoint
// for creating memberships related to a set of Keyrings.
type KeyringMemberClientV1 struct {
	client RoundTripper
}

// Post sends a creation requests for a set of KeyringMember objects to the
// registry.
func (k *KeyringMemberClientV1) Post(ctx context.Context, members []envelope.KeyringMemberV1) ([]envelope.KeyringMemberV1, error) {

	req, err := k.client.NewRequest("POST", "/keyring-members", nil, members)
	if err != nil {
		log.Printf("Error creating POST /keyring-members request: %s", err)
		return nil, err
	}

	resp := []envelope.KeyringMemberV1{}
	_, err = k.client.Do(ctx, req, &resp)
	if err != nil {
		log.Printf("Error performing POST /keyring-members request: %s", err)
		return nil, err
	}

	return resp, err
}

// KeyringMembersClient represents the `/keyring/:id/members` registry endpoint
// for creating memberships in a keyring.
type KeyringMembersClient struct {
	client RoundTripper
}

// Post sends a creation requests for a set of KeyringMember objects to the
// registry.
func (k *KeyringMembersClient) Post(ctx context.Context, member KeyringMember) error {
	keyringID := member.Member.Body.KeyringID
	members := []KeyringMember{member}
	req, err := k.client.NewRequest("POST", "/keyrings/"+keyringID.String()+"/members", nil, members)
	if err != nil {
		log.Printf("Error creating POST /keyring/:id/members request: %s", err)
		return err
	}

	_, err = k.client.Do(ctx, req, nil)
	if err != nil {
		log.Printf("Error performing POST /keyring/:id/members request: %s", err)
		return err
	}

	return err
}
