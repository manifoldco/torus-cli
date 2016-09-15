package registry

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/pathexp"
)

// CredentialGraphClient represents the `/credentialgraph` registry endpoint,
// user for retrieving keyrings, keyring members, and credentials associated
// with claims.
type CredentialGraphClient struct {
	client *Client
}

// CredentialGraph is the shared interface between different credential graph
// versions
type CredentialGraph interface {
	KeyringSection
	GetCredentials() []envelope.Signed
}

// CredentialGraphV1 represents a Keyring, it's members, and associated
// Credentials.
type CredentialGraphV1 struct {
	KeyringSectionV1
	Credentials []envelope.Signed `json:"credentials"`
}

// GetCredentials returns the Credentials objects in this CredentialGraph
func (c *CredentialGraphV1) GetCredentials() []envelope.Signed {
	return c.Credentials
}

// CredentialGraphV2 represents a Keyring, it's members, and associated
// Credentials.
type CredentialGraphV2 struct {
	KeyringSectionV2
	Credentials []envelope.Signed `json:"credentials"`
}

// GetCredentials returns the Credentials objects in this CredentialGraph
func (c *CredentialGraphV2) GetCredentials() []envelope.Signed {
	return c.Credentials
}

// Post creates a new CredentialGraph on the registry.
//
// The CredentialGraph includes the keyring, it's members, and credentials.
func (c *CredentialGraphClient) Post(ctx context.Context, t *CredentialGraph) (*CredentialGraphV2, error) {
	req, err := c.client.NewRequest("POST", "/credentialgraph", nil, t)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := CredentialGraphV2{}
	_, err = c.client.Do(ctx, req, &resp)
	if err != nil {
		log.Printf("Failed to create credential graph: %s", err)
		return nil, err
	}

	return &resp, nil
}

// List returns back all segments of the CredentialGraph (Keyring, Keyring
// Members, and Credentials) that match the given name, path, or path
// expression.
func (c *CredentialGraphClient) List(ctx context.Context, name, path string,
	pathExp *pathexp.PathExp, ownerID *identity.ID) ([]CredentialGraph, error) {

	query := url.Values{}

	if path != "" && pathExp != nil {
		return nil, errors.New("cannot provide path and pathexp at the same time")
	}
	if path != "" {
		query.Set("path", path)
	}
	if pathExp != nil {
		query.Set("pathexp", pathExp.String())
	}
	if name != "" {
		query.Set("name", name)
	}
	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	req, err := c.client.NewRequest("GET", "/credentialgraph", &query, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := []struct {
		Keyring     *envelope.Signed  `json:"keyring"`
		Members     json.RawMessage   `json:"members"`
		Credentials []envelope.Signed `json:"credentials"`
		Claims      []envelope.Signed `json:"claims"`
	}{}

	_, err = c.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	converted := make([]CredentialGraph, len(resp))
	for i, g := range resp {
		if g.Keyring.Version == 1 {
			c := CredentialGraphV1{
				KeyringSectionV1: KeyringSectionV1{Keyring: g.Keyring},
				Credentials:      g.Credentials,
			}
			err := json.Unmarshal(g.Members, &c.Members)
			if err != nil {
				return nil, err
			}
			converted[i] = &c
		} else {
			c := CredentialGraphV2{
				KeyringSectionV2: KeyringSectionV2{
					Keyring: g.Keyring,
					Claims:  g.Claims,
				},
				Credentials: g.Credentials,
			}
			err := json.Unmarshal(g.Members, &c.Members)
			if err != nil {
				return nil, err
			}
			converted[i] = &c
		}
	}

	return converted, nil
}
