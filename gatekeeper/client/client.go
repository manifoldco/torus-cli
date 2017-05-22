// Package client provides the Gatekeeper bootstrap API
package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/registry"
)

const (
	gatekeeperAPIVersion = "v0"
	clientTimeout        = time.Minute
)

type clientRoundTripper struct {
	registry.DefaultRequestDoer // TODO: Could be a generic RequestDoer?
}

// Client is the Gatekeeper bootstrapping client
type Client struct {
	rt *clientRoundTripper
}

// NewClient returns a new client to a Gatekeeper host that can bootstrap this machine
func NewClient(host string) *Client {
	return &Client{
		rt: &clientRoundTripper{
			DefaultRequestDoer: registry.DefaultRequestDoer{
				Client: &http.Client{
					Timeout: clientTimeout,
				},
				Host: host,
			},
		},
	}
}

// Bootstrap bootstraps the machine with Gatekeeper
func (c *Client) Bootstrap(bootreq interface{}) (*apitypes.BootstrapResponse, error) {
	path := fmt.Sprintf("%s/%s", gatekeeperAPIVersion, "machine")
	req, err := c.rt.NewRequest("POST", path, nil, bootreq)
	if err != nil {
		return nil, err
	}

	var bootresp apitypes.BootstrapResponse
	resp, err := c.rt.Do(context.Background(), req, &bootresp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &bootresp, nil
}
