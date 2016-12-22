package api

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// upstreamKeypairsClient makes proxied requests to the registry's keypairs
// endpoints
type upstreamKeypairsClient struct {
	client RoundTripper
}

// KeypairsClient makes requests to the registry's and daemon's keypairs
// endpoints
type KeypairsClient struct {
	upstreamKeypairsClient
	client *Client
}

func newKeypairsClient(c *Client) *KeypairsClient {
	return &KeypairsClient{upstreamKeypairsClient{c}, c}
}

// KeypairResult is the payload returned for a keypair object
type KeypairResult struct {
	PublicKey  *envelope.PublicKey  `json:"public_key"`
	PrivateKey *envelope.PrivateKey `json:"private_key"`
	Claims     []envelope.Claim     `json:"claims"`
}

// Revoked returns a bool indicating if any revocation claims exist against this
// KeypairResult's keypair.
func (k *KeypairResult) Revoked() bool {
	for _, claim := range k.Claims {
		if claim.Body.ClaimType == primitive.RevocationClaimType {
			return true
		}
	}

	return false
}

type keypairsRequest struct {
	OrgID *identity.ID `json:"org_id"`
}

// Generate generates new keypairs for the user in the given org.
func (k *KeypairsClient) Generate(ctx context.Context, orgID *identity.ID,
	output ProgressFunc) error {

	kpr := keypairsRequest{OrgID: orgID}

	req, reqID, err := k.client.NewDaemonRequest("POST", "/keypairs/generate", nil, &kpr)
	if err != nil {
		return err
	}

	_, err = k.client.DoWithProgress(ctx, req, nil, reqID, output)
	return err
}

// List retrieves relevant keypairs by orgID
func (k *upstreamKeypairsClient) List(ctx context.Context, orgID *identity.ID) ([]KeypairResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, err := k.client.NewRequest("GET", "/keypairs", v, nil)
	if err != nil {
		return nil, err
	}

	var keypairs []KeypairResult
	_, err = k.client.Do(ctx, req, &keypairs)
	if err != nil {
		return nil, err
	}

	return keypairs, nil
}

// Revoke revokes the existing keypairs for the user in the given org.
func (k *KeypairsClient) Revoke(ctx context.Context, orgID *identity.ID, output ProgressFunc) error {
	kpr := keypairsRequest{OrgID: orgID}

	req, reqID, err := k.client.NewDaemonRequest("POST", "/keypairs/revoke", nil, &kpr)
	if err != nil {
		return err
	}

	_, err = k.client.DoWithProgress(ctx, req, nil, reqID, output)
	return err
}
