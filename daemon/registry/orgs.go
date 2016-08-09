package registry

import (
	"log"
	"net/url"

	"github.com/arigatomachine/cli/daemon/envelope"
)

// Orgs represents the `/orgs` registry endpoint, used for accessing
// organizations stored in Arigato.
type Orgs struct {
	client *Client
}

// List returns all organizations that match the given name.
func (o *Orgs) List(name string) ([]envelope.Unsigned, error) {
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
	_, err = o.client.Do(req, &orgs)
	if err != nil {
		log.Printf("Error performing api request: %s", err)
		return nil, err
	}

	return orgs, nil
}
