package registry

import (
	"log"
	"net/url"

	"github.com/arigatomachine/cli/daemon/envelope"
)

type Orgs struct {
	client *Client
}

func (o *Orgs) List(name string) ([]envelope.Unsigned, error) {
	v := url.Values{}

	if name != "" {
		v.Set("name", name)
	}

	req, err := o.client.NewRequest("GET", "/orgs", &v, nil)
	if err != nil {
		log.Printf("Error making api request: %s", err)
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
