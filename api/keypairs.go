package api

import (
	"context"
	"net/url"

	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
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

// List retrieves relevant keypairs by orgID
func (o *KeypairsClient) List(ctx context.Context, orgID *identity.ID) ([]KeypairResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, _, err := o.client.NewRequest("GET", "/keypairs", v, nil, true)
	if err != nil {
		return nil, err
	}

	var keypairs []KeypairResult
	_, err = o.client.Do(ctx, req, &keypairs, nil, nil)
	if err != nil {
		return nil, err
	}

	return keypairs, nil
}
