package registry

import (
	"context"

	"github.com/manifoldco/torus-cli/envelope"
)

// Credentials represents the `/credentials` registry endpoint, used for
// accessing encrypted credentials/secrets.
type Credentials struct {
	client RoundTripper
}

// Create creates the provided credential in the registry.
func (c *Credentials) Create(ctx context.Context, creds []envelope.CredentialInf) ([]*envelope.Credential, error) {
	resp := []*envelope.Credential{}
	err := c.client.RoundTrip(ctx, "POST", "/credentials", nil, creds, &resp)
	return resp, err
}
