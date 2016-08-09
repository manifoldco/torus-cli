package registry

import (
	"log"

	"github.com/arigatomachine/cli/daemon/envelope"
)

// Credentials represents the `/credentials` registry endpoint, used for
// accessing encrypted credentials/secrets.
type Credentials struct {
	client *Client
}

// Create creates the provided credential in the registry.
func (c *Credentials) Create(credential *envelope.Signed) (*envelope.Signed, error) {
	req, err := c.client.NewRequest("POST", "/credentials", nil, credential)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := &envelope.Signed{}
	_, err = c.client.Do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
