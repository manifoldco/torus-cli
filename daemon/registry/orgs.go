package registry

import (
	"context"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
)

// Orgs represents the `/orgs` registry endpoint, used for accessing
// organizations stored in Torus.
type Orgs struct {
	client *Client
}

// List returns all organizations that match the given name.
func (o *Orgs) List(ctx context.Context, name string) ([]envelope.Unsigned, error) {
	v := url.Values{}

	if name != "" {
		v.Set("name", name)
	}

	req, err := o.client.NewRequest("GET", "/orgs", &v, nil)
	if err != nil {
		log.Printf("Error building GET /orgs api request: %s", err)
		return nil, err
	}

	orgs := []envelope.Unsigned{}
	_, err = o.client.Do(ctx, req, &orgs)
	if err != nil {
		log.Printf("Error performing api request: %s", err)
		return nil, err
	}

	return orgs, nil
}

// Get returns the organization with the given ID.
func (o *Orgs) Get(ctx context.Context, orgID *identity.ID) (*envelope.Unsigned, error) {
	req, err := o.client.NewRequest("GET", "/orgs/"+orgID.String(), nil, nil)
	if err != nil {
		log.Printf("Error building GET /orgs api request: %s", err)
		return nil, err
	}

	org := envelope.Unsigned{}
	_, err = o.client.Do(ctx, req, &org)
	if err != nil {
		log.Printf("Error performing api request: %s", err)
		return nil, err
	}

	return &org, nil
}
