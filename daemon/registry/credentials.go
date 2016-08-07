package registry

import (
	"errors"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/daemon/envelope"
)

// Credentials represents the `/credentials` registry endpoint, used for
// accessing encrypted credentials/secrets.
type Credentials struct {
	client *Client
}

// Create creates the provided credential in the registry.
func (c *Credentials) Create(credential *envelope.Unsigned) (*envelope.Unsigned, error) {
	req, err := c.client.NewRequest("POST", "/credentials", nil, credential)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := &envelope.Unsigned{}
	_, err = c.client.Do(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// List returns all credentials that match the given  arguments.
func (c *Credentials) List(name, path, pathexp string) ([]envelope.Unsigned, error) {
	query := url.Values{}

	if path != "" && pathexp != "" {
		return nil, errors.New("cannot provide cpath and cpathexp at same time")
	}
	if path != "" {
		query.Set("path", path)
	}
	if pathexp != "" {
		query.Set("pathexp", pathexp)
	}
	if name != "" {
		query.Set("name", name)
	}

	req, err := c.client.NewRequest("GET", "/credentials", &query, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := []envelope.Unsigned{}
	_, err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
