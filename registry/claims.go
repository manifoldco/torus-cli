package registry

import (
	"context"

	"github.com/manifoldco/torus-cli/envelope"
)

// ClaimsClient represents the `/claims` registry endpoint for making claims
// against keypairs. Claims can either be a signature or a revocation.
type ClaimsClient struct {
	client RoundTripper
}

// Create creates a a new signed claim on the server
func (c ClaimsClient) Create(ctx context.Context, claim *envelope.Claim) (*envelope.Claim, error) {
	resp := &envelope.Claim{}
	err := c.client.RoundTrip(ctx, "POST", "/claims", nil, claim, &resp)
	return resp, err
}
