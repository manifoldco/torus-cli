package registry

import (
	"context"

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
	resp := []envelope.KeyringMemberV1{}
	err := k.client.RoundTrip(ctx, "POST", "/keyring-members", nil, members, &resp)
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
	return k.client.RoundTrip(ctx, "POST", "/keyrings/"+keyringID.String()+"/members", nil, members, nil)
}
