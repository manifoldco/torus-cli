package api

import (
	"context"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
)

// CredentialsClient provides access to unencrypted credentials for viewing,
// and encrypts credentials when setting.
type CredentialsClient struct {
	client *Client
}

// Get returns all credentials at the given path.
func (c *CredentialsClient) Get(ctx context.Context, path string) ([]apitypes.CredentialEnvelope, error) {
	v := &url.Values{}
	v.Set("path", path)

	req, _, err := c.client.NewRequest("GET", "/credentials", v, nil, false)
	if err != nil {
		return nil, err
	}

	var creds []apitypes.CredentialEnvelope
	_, err = c.client.Do(ctx, req, &creds, nil, nil)
	return creds, err
}
