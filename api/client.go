// Package api provides the daemon API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/config"
)

// Client exposes the daemon API.
type Client struct {
	client *http.Client

	Orgs        *OrgsClient
	Users       *UsersClient
	Profiles    *ProfilesClient
	Teams       *TeamsClient
	Memberships *MembershipsClient
	Invites     *InvitesClient
	Session     *SessionClient
	Version     *VersionClient
}

// NewClient returns a new Client.
func NewClient(cfg *config.Config) *Client {
	c := &Client{
		client: &http.Client{
			Transport: &http.Transport{
				Dial: func(network, address string) (net.Conn, error) {
					return net.Dial("unix", cfg.SocketPath)
				},
			},
		},
	}

	c.Orgs = &OrgsClient{client: c}
	c.Users = &UsersClient{client: c}
	c.Profiles = &ProfilesClient{client: c}
	c.Teams = &TeamsClient{client: c}
	c.Memberships = &MembershipsClient{client: c}
	c.Invites = &InvitesClient{client: c}
	c.Session = &SessionClient{client: c}
	c.Version = &VersionClient{client: c}

	return c
}

// NewRequest constructs a new http.Request, with a body containing the json
// representation of body, if provided.
func (c *Client) NewRequest(method, path string, query *url.Values,
	body interface{}, proxied bool) (*http.Request, error) {

	b := &bytes.Buffer{}
	if body != nil {
		enc := json.NewEncoder(b)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	if query == nil {
		query = &url.Values{}
	}

	version := "v1"
	if proxied {
		version = "proxy"
	}

	fullPath := "http://localhost/" + version + path + "?" + query.Encode()
	req, err := http.NewRequest(method, fullPath, b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", "localhost")
	req.Header.Set("Content-type", "application/json")

	return req, nil
}

// Do executes an http.Request, populating v with the JSON response
// on success.
//
// If the request errors with a JSON formatted response body, it will be
// unmarshaled into the returned error.
func (c *Client) Do(ctx context.Context, r *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = checkResponseCode(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(v)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func checkResponseCode(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	rErr := &apitypes.Error{StatusCode: r.StatusCode}
	if r.ContentLength != 0 {
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(rErr)
		if err != nil {
			return errors.New("Malformed error response from daemon.")
		}

		return rErr
	}

	return errors.New("Error from daemon. Check status code.")
}
