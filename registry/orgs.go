package registry

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// Orgs represents the `/orgs` registry endpoint, used for accessing
// organizations stored in Torus.
type Orgs struct {
	client RoundTripper
}

// Get returns the organization with the given ID.
func (o *Orgs) Get(ctx context.Context, orgID *identity.ID) (*envelope.Org, error) {
	req, err := o.client.NewRequest("GET", "/orgs/"+orgID.String(), nil, nil)
	if err != nil {
		log.Printf("Error building GET /orgs api request: %s", err)
		return nil, err
	}

	org := envelope.Org{}
	_, err = o.client.Do(ctx, req, &org)
	if err != nil {
		log.Printf("Error performing api request: %s", err)
		return nil, err
	}

	return &org, nil
}
