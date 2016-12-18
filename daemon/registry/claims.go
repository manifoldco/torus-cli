package registry

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/envelope"
)

// ClaimsClient represents the `/claims` registry endpoint for making claims
// against keypairs. Claims can either be a signature or a revocation.
type ClaimsClient struct {
	client *Client
}

// Create creates a a new signed claim on the server
func (c ClaimsClient) Create(ctx context.Context, claim *envelope.Claim) (*envelope.Claim, error) {
	req, err := c.client.NewRequest("POST", "/claims", nil, claim)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := &envelope.Claim{}
	_, err = c.client.Do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
