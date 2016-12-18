package registry

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
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
	GetCredentials() []envelope.CredentialInf
}

// CredentialGraphV1 represents a Keyring, it's members, and associated
// Credentials.
type CredentialGraphV1 struct {
	KeyringSectionV1
	Credentials []envelope.CredentialInf `json:"credentials"`
}

// GetCredentials returns the Credentials objects in this CredentialGraph
func (c *CredentialGraphV1) GetCredentials() []envelope.CredentialInf {
	return c.Credentials
}

// CredentialGraphV2 represents a Keyring, it's members, and associated
// Credentials.
type CredentialGraphV2 struct {
	KeyringSectionV2
	Credentials []envelope.CredentialInf `json:"credentials"`
}

// GetCredentials returns the Credentials objects in this CredentialGraph
func (c *CredentialGraphV2) GetCredentials() []envelope.CredentialInf {
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
func (c *CredentialGraphClient) List(ctx context.Context, path string,
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
	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	return c.getGraph(ctx, query)
}

// Search returns back all segments of the CredentialGraph (Keyring, Keyring
// Members, and Credentials) that are contained within the given loose path
// expression. It is loose in that it can have * for projects.
func (c *CredentialGraphClient) Search(ctx context.Context, pathExp string,
	ownerID *identity.ID) ([]CredentialGraph, error) {

	query := url.Values{}

	query.Set("pathexp", pathExp)
	query.Set("owner_id", ownerID.String())
	query.Set("mode", "contains")

	return c.getGraph(ctx, query)
}

func (c *CredentialGraphClient) getGraph(ctx context.Context, query url.Values) ([]CredentialGraph, error) {
	req, err := c.client.NewRequest("GET", "/credentialgraph", &query, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := []struct {
		Keyring     *envelope.Signed              `json:"keyring"`
		Members     json.RawMessage               `json:"members"`
		Credentials []envelope.Signed             `json:"credentials"`
		Claims      []envelope.KeyringMemberClaim `json:"claims"`
	}{}

	_, err = c.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	converted := make([]CredentialGraph, len(resp))
	for i, g := range resp {
		creds := make([]envelope.CredentialInf, len(g.Credentials))
		for i, ec := range g.Credentials {
			switch ec.Body.(type) {
			case *primitive.CredentialV1:
				creds[i] = &envelope.CredentialV1{
					ID:        ec.ID,
					Version:   ec.Version,
					Signature: ec.Signature,
					Body:      ec.Body.(*primitive.CredentialV1),
				}
			case *primitive.Credential:
				creds[i] = &envelope.Credential{
					ID:        ec.ID,
					Version:   ec.Version,
					Signature: ec.Signature,
					Body:      ec.Body.(*primitive.Credential),
				}
			}
		}

		if g.Keyring.Version == 1 {
			kre := &envelope.KeyringV1{
				ID:        g.Keyring.ID,
				Version:   g.Keyring.Version,
				Signature: g.Keyring.Signature,
				Body:      g.Keyring.Body.(*primitive.KeyringV1),
			}

			c := CredentialGraphV1{
				KeyringSectionV1: KeyringSectionV1{
					Keyring: kre,
				},
				Credentials: creds,
			}
			err := json.Unmarshal(g.Members, &c.Members)
			if err != nil {
				return nil, err
			}
			converted[i] = &c
		} else {
			kre := &envelope.Keyring{
				ID:        g.Keyring.ID,
				Version:   g.Keyring.Version,
				Signature: g.Keyring.Signature,
				Body:      g.Keyring.Body.(*primitive.Keyring),
			}

			c := CredentialGraphV2{
				KeyringSectionV2: KeyringSectionV2{
					Keyring: kre,
					Claims:  g.Claims,
				},
				Credentials: creds,
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
