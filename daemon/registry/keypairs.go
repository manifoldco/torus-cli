package registry

import (
	"log"

	"github.com/arigatomachine/cli/daemon/envelope"
)

type keyPairsCreateRequest struct { // Its the response, too!
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
func (k *KeyPairs) Post(pubKey, privKey, claim *envelope.Signed) (
	*envelope.Signed, *envelope.Signed, []envelope.Signed, error) {

	req, err := k.client.NewRequest("POST", "/keypairs", nil,
		keyPairsCreateRequest{
			PublicKey:  pubKey,
			PrivateKey: privKey,
			Claims:     []envelope.Signed{*claim},
		})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, nil, nil, err
	}

	resp := keyPairsCreateRequest{}
	_, err = k.client.Do(req, &resp)
	if err != nil {
		log.Printf("Failed to create signing keypair: %s", err)
		return nil, nil, nil, err
	}

	return resp.PublicKey, resp.PrivateKey, resp.Claims, nil
}
