// Package client provides the Gatekeeper bootstrap API
package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/registry"
)

const (
	gatekeeperAPIVersion = "v0"
	clientTimeout        = time.Minute
)

type clientRoundTripper struct {
	// TODO: Could abstract the registry RequestDoer into package other than
	// registry to avoid confusion (since we're not talking to the registry here)
	registry.DefaultRequestDoer
}

// Client is the Gatekeeper bootstrapping client
type Client struct {
	rt *clientRoundTripper
}

// NewClient returns a new client to a Gatekeeper host that can bootstrap this machine
func NewClient(host string) *Client {
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
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
func (c *Client) Bootstrap(provider string, bootreq interface{}) (*apitypes.BootstrapResponse, error) {
	path := fmt.Sprintf("%s/%s/%s", gatekeeperAPIVersion, "machine", provider)
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
