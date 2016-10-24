package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
)

// CredentialsClient provides access to unencrypted credentials for viewing,
// and encrypts credentials when setting.
type CredentialsClient struct {
	client *Client
}

// Search returns all credentials at the given pathexp.
func (c *CredentialsClient) Search(ctx context.Context, pathexp string) ([]apitypes.CredentialEnvelope, error) {
	v := &url.Values{}
	v.Set("pathexp", pathexp)

	req, _, err := c.client.NewRequest("GET", "/credentials", v, nil, false)
	if err != nil {
		return nil, err
	}

	resp := []apitypes.CredentialResp{}

	_, err = c.client.Do(ctx, req, &resp, nil, nil)
	if err != nil {
		return nil, err
	}

	creds := make([]apitypes.CredentialEnvelope, len(resp))
	for i, c := range resp {
		v, err := createEnvelopeFromResp(c)
		if err != nil {
			return nil, err
		}
		creds[i] = *v
	}

	return creds, err
}

// Get returns all credentials at the given path.
func (c *CredentialsClient) Get(ctx context.Context, path string) ([]apitypes.CredentialEnvelope, error) {
	v := &url.Values{}
	v.Set("path", path)

	req, _, err := c.client.NewRequest("GET", "/credentials", v, nil, false)
	if err != nil {
		return nil, err
	}

	resp := []apitypes.CredentialResp{}

	_, err = c.client.Do(ctx, req, &resp, nil, nil)
	if err != nil {
		return nil, err
	}

	creds := make([]apitypes.CredentialEnvelope, len(resp))
	for i, c := range resp {
		v, err := createEnvelopeFromResp(c)
		if err != nil {
			return nil, err
		}
		creds[i] = *v
	}

	return creds, err
}

// Create creates the given credential
func (c *CredentialsClient) Create(ctx context.Context, cred *apitypes.Credential,
	progress *ProgressFunc) (*apitypes.CredentialEnvelope, error) {

	env := apitypes.CredentialEnvelope{Version: 2, Body: cred}
	req, reqID, err := c.client.NewRequest("POST", "/credentials", nil, &env, false)
	if err != nil {
		return nil, err
	}

	resp := apitypes.CredentialResp{}
	_, err = c.client.Do(ctx, req, &resp, &reqID, progress)
	if err != nil {
		return nil, err
	}

	out, err := createEnvelopeFromResp(resp)
	return out, err
}

func createEnvelopeFromResp(c apitypes.CredentialResp) (*apitypes.CredentialEnvelope, error) {
	var envelope apitypes.CredentialEnvelope
	var cBody apitypes.Credential
	switch c.Version {
	case 1:
		cBodyV1 := apitypes.BaseCredential{}
		err := json.Unmarshal(c.Body, &cBodyV1)
		if err != nil {
			return nil, err
		}

		cBody = &cBodyV1
	case 2:
		cBodyV2 := apitypes.CredentialV2{}
		err := json.Unmarshal(c.Body, &cBodyV2)
		if err != nil {
			return nil, err
		}

		cBody = &cBodyV2
	default:
		return nil, errors.New("Unknown credential version")
	}

	envelope = apitypes.CredentialEnvelope{
		ID:      c.ID,
		Version: c.Version,
		Body:    &cBody,
	}
	return &envelope, nil
}
