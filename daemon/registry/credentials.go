package registry

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/envelope"
)

// Credentials represents the `/credentials` registry endpoint, used for
// accessing encrypted credentials/secrets.
type Credentials struct {
	client *Client
}

// Create creates the provided credential in the registry.
func (c *Credentials) Create(ctx context.Context, credential *envelope.Credential) (*envelope.Credential, error) {
	req, err := c.client.NewRequest("POST", "/credentials", nil, credential)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := &envelope.Credential{}
	_, err = c.client.Do(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
