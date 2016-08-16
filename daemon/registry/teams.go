package registry

import (
	"context"
	"errors"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
)

// TeamsClient represents the `/teams` registry endpoint, used for
// accessing teams stored in Arigato.
type TeamsClient struct {
	client *Client
}

// List returns all teams for an organization
func (t *TeamsClient) List(orgID *identity.ID) ([]envelope.Unsigned, error) {
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
	_, err = t.client.Do(context.TODO(), req, &teams)
	if err != nil {
		log.Printf("Error performing GET /teams request: %s", err)
		return nil, err
	}

	return teams, nil
}
