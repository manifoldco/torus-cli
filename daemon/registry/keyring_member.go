package registry

import (
	"log"

	"github.com/arigatomachine/cli/daemon/envelope"
)

// KeyringMemberClient represents the `/keyring-members` registry end point for
// accessand creating memberships related to a set of Keyrings.
type KeyringMemberClient struct {
	client *Client
}

// Post sends a creation requests for a set of KeyringMember objects to the
// registry.
func (k *KeyringMemberClient) Post(members []envelope.Signed) ([]envelope.Signed, error) {

	req, err := k.client.NewRequest("POST", "/keyring-members", nil, members)
	if err != nil {
		log.Printf("Error creating POST /keyring-members request: %s", err)
		return nil, err
	}

	resp := []envelope.Signed{}
	_, err = k.client.Do(req, &resp)
	if err != nil {
		log.Printf("Error performing POST /keyring-members request: %s", err)
		return nil, err
	}

	return resp, err
}
