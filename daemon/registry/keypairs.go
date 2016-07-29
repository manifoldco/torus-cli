package registry

import (
	"log"

	"github.com/arigatomachine/cli/daemon/envelope"
)

type KeyPairs struct {
	client *Client
}

func (k *KeyPairs) Post(pubKey, privKey, claim *envelope.Signed) (
	*envelope.Signed, *envelope.Signed, []envelope.Signed, error) {

	req, err := k.client.NewRequest("POST", "/keypairs",
		KeyPairsCreateRequest{
			PublicKey:  pubKey,
			PrivateKey: privKey,
			Claims:     []envelope.Signed{*claim},
		})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, nil, nil, err
	}

	resp := KeyPairsCreateRequest{}
	_, err = k.client.Do(req, &resp)
	if err != nil {
		log.Printf("Failed to create signing keypair: %s", err)
		return nil, nil, nil, err
	}

	return resp.PublicKey, resp.PrivateKey, resp.Claims, nil
}
