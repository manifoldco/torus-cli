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

// Create generates new keypairs for the user in the given org.
func (k *KeyPairsClient) Create(ctx context.Context, orgID *identity.ID, output ProgressFunc) error {
	return k.worker(ctx, "generate", orgID, output)
}

// Revoke revokes the existing keypairs for the user in the given org.
func (k *KeyPairsClient) Revoke(ctx context.Context, orgID *identity.ID, output ProgressFunc) error {
	return k.worker(ctx, "revoke", orgID, output)
}

func (k *KeyPairsClient) worker(ctx context.Context, action string, orgID *identity.ID, output ProgressFunc) error {
	kpr := keyPairsRequest{OrgID: orgID}

	req, reqID, err := k.client.NewDaemonRequest("POST", "/keypairs/"+action, nil, &kpr)
	if err != nil {
		return err
	}

	_, err = k.client.DoWithProgress(ctx, req, nil, reqID, output)
	return err
}
