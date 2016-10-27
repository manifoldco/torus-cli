package registry

import (
	"context"
	"errors"
	"log"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// TeamsClient represents the `/teams` registry endpoint, used for
// accessing teams stored in Torus.
type TeamsClient struct {
	client *Client
}

// List returns all teams for an organization
func (t *TeamsClient) List(ctx context.Context, orgID *identity.ID) ([]envelope.Unsigned, error) {
	if orgID == nil {
		return nil, errors.New("must provide org id")
	}

	v := &url.Values{}
	v.Set("org_id", orgID.String())

	req, err := t.client.NewRequest("GET", "/teams", v, nil)
	if err != nil {
		log.Printf("Error building GET /teams request: %s", err)
		return nil, err
	}

	teams := []envelope.Unsigned{}
	_, err = t.client.Do(ctx, req, &teams)
	if err != nil {
		log.Printf("Error performing GET /teams request: %s", err)
		return nil, err
	}

	return teams, nil
}
