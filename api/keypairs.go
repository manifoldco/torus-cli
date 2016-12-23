package api

import (
	"context"

	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/registry"
)

// KeyPairsClient makes requests to the registry's and daemon's keypairs
// endpoints
type KeyPairsClient struct {
	*registry.KeyPairsClient
	client *apiRoundTripper
}

func newKeyPairsClient(upstream *registry.KeyPairsClient, rt *apiRoundTripper) *KeyPairsClient {
	return &KeyPairsClient{upstream, rt}
}

type keyPairsRequest struct {
	OrgID *identity.ID `json:"org_id"`
}

// Generate generates new keypairs for the user in the given org.
func (k *KeyPairsClient) Generate(ctx context.Context, orgID *identity.ID,
	output ProgressFunc) error {

	kpr := keyPairsRequest{OrgID: orgID}

	req, reqID, err := k.client.NewDaemonRequest("POST", "/keypairs/generate", nil, &kpr)
	if err != nil {
		return err
	}

	_, err = k.client.DoWithProgress(ctx, req, nil, reqID, output)
	return err
}

// Revoke revokes the existing keypairs for the user in the given org.
func (k *KeyPairsClient) Revoke(ctx context.Context, orgID *identity.ID, output ProgressFunc) error {
	kpr := keyPairsRequest{OrgID: orgID}

	req, reqID, err := k.client.NewDaemonRequest("POST", "/keypairs/revoke", nil, &kpr)
	if err != nil {
		return err
	}

	_, err = k.client.DoWithProgress(ctx, req, nil, reqID, output)
	return err
}
