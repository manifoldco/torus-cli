package api

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// KeypairsClient makes proxied requests to the registry's keypairs endpoints
type KeypairsClient struct {
	client *Client
}

// KeypairResult is the payload returned for a keypair object
type KeypairResult struct {
	PublicKey *struct {
		ID   *identity.ID         `json:"id"`
		Body *primitive.PublicKey `json:"body"`
	} `json:"public_key"`
	PrivateKey *struct {
		ID   *identity.ID         `json:"id"`
		Body *primitive.PublicKey `json:"body"`
	} `json:"private_key"`
	Claims *[]struct {
		ID   *identity.ID     `json:"id"`
		Body *primitive.Claim `json:"claims"`
	}
}

type keypairsGenerateRequest struct {
	OrgID *identity.ID `json:"org_id"`
}

// Generate generates new keypairs for the user in the given org.
func (k *KeypairsClient) Generate(ctx context.Context, orgID *identity.ID,
	output *ProgressFunc) error {

	kpgr := keypairsGenerateRequest{OrgID: orgID}

	req, reqID, err := k.client.NewRequest("POST", "/keypairs/generate", nil, &kpgr, false)
	if err != nil {
		return err
	}

	_, err = k.client.Do(ctx, req, nil, &reqID, output)
	return err
}

// List retrieves relevant keypairs by orgID
func (k *KeypairsClient) List(ctx context.Context, orgID *identity.ID) ([]KeypairResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, _, err := k.client.NewRequest("GET", "/keypairs", v, nil, true)
	if err != nil {
		return nil, err
	}

	var keypairs []KeypairResult
	_, err = k.client.Do(ctx, req, &keypairs, nil, nil)
	if err != nil {
		return nil, err
	}

	return keypairs, nil
}
