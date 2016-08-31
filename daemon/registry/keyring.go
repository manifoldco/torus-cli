package registry

import (
	"context"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/identity"

	"github.com/arigatomachine/cli/daemon/envelope"
)

// KeyringClient represents the `/keyrings` registry end point for accessing
// keyrings the user or machine belong too.
type KeyringClient struct {
	client *Client
}

// KeyringSection represents a section of the CredentialTree only pertaining to
// a keyring and it's membership.
type KeyringSection struct {
	Keyring *envelope.Signed  `json:"keyring"`
	Members []envelope.Signed `json:"members"`
}

// List retrieves an array of KeyringSections from the registry.
func (k *KeyringClient) List(ctx context.Context, orgID *identity.ID,
	ownerID *identity.ID) ([]KeyringSection, error) {

	query := &url.Values{}

	if orgID != nil {
		query.Set("org_id", orgID.String())
	}
	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	req, err := k.client.NewRequest("GET", "/keyrings", query, nil)
	if err != nil {
		log.Printf("Error building http request for GET /keyrings: %s", err)
		return nil, err
	}

	resp := []KeyringSection{}
	_, err = k.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
