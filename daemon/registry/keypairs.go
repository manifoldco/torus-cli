package registry

import (
	"context"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
)

// ClaimedKeyPair contains a public/private keypair, and all the Claims made
// against it (system and user signatures).
type ClaimedKeyPair struct {
	PublicKey  *envelope.Signed  `json:"public_key"`
	PrivateKey *envelope.Signed  `json:"private_key"`
	Claims     []envelope.Signed `json:"claims"`
}

// KeyPairs represents the `/keypairs` registry endpoint, used for accessing
// users' signing and encryption keypairs.
type KeyPairs struct {
	client *Client
}

// Post creates a new keypair on the registry.
//
// The keypair includes the user's public key, private key, and a self-signed
// claim on the public key.
//
// keys may be either signing or encryption keys.
func (k *KeyPairs) Post(ctx context.Context, pubKey, privKey,
	claim *envelope.Signed) (*envelope.Signed, *envelope.Signed, []envelope.Signed, error) {

	req, err := k.client.NewRequest("POST", "/keypairs", nil,
		ClaimedKeyPair{
			PublicKey:  pubKey,
			PrivateKey: privKey,
			Claims:     []envelope.Signed{*claim},
		})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, nil, nil, err
	}

	resp := ClaimedKeyPair{}
	_, err = k.client.Do(ctx, req, &resp)
	if err != nil {
		log.Printf("Failed to create signing keypair: %s", err)
		return nil, nil, nil, err
	}

	return resp.PublicKey, resp.PrivateKey, resp.Claims, nil
}

// List returns all KeyPairs for the logged in user in the given, or all orgs
// if orgID is nil.
func (k *KeyPairs) List(ctx context.Context, orgID *identity.ID) ([]ClaimedKeyPair, error) {
	query := &url.Values{}
	if orgID != nil {
		query.Set("org_id", orgID.String())
	}

	req, err := k.client.NewRequest("GET", "/keypairs", query, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := []ClaimedKeyPair{}
	_, err = k.client.Do(ctx, req, &resp)
	if err != nil {
		log.Printf("Failed to retrieve keypairs: %s", err)
		return nil, err
	}

	return resp, nil
}
