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

	"github.com/donovanhide/eventsource"
	"github.com/satori/go.uuid"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
)

// RoundTripper is the interface used to construct and send requests to
// the torus registry.
type RoundTripper interface {
	NewRequest(method, path string, query *url.Values, body interface{}) (*http.Request, error)
	Do(ctx context.Context, r *http.Request, v interface{}) (*http.Response, error)
}

type proxy struct {
	c *Client
}

func (p proxy) NewRequest(method string, path string, query *url.Values, body interface{}) (*http.Request, error) {
	req, _, err := p.c.NewRequest(method, path, query, body, true)
	return req, err
}

func (p proxy) Do(ctx context.Context, r *http.Request, v interface{}) (*http.Response, error) {
	return p.c.Do(ctx, r, v, nil, nil)
}

// Client exposes the daemon API.
type Client struct {
	client *http.Client

	Orgs         *OrgsClient
	Users        *UsersClient
	Machines     *MachinesClient
	Profiles     *ProfilesClient
	Teams        *TeamsClient
	Memberships  *MembershipsClient
	Invites      *InvitesClient
	Keypairs     *KeypairsClient
	Session      *SessionClient
	Services     *ServicesClient
	Policies     *PoliciesClient
	Environments *EnvironmentsClient
	Projects     *ProjectsClient
	Credentials  *CredentialsClient
	Worklog      *WorklogClient
	Version      *VersionClient
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

	p := &proxy{c: c}

	c.Orgs = &OrgsClient{client: p}
	c.Users = &UsersClient{client: c}
	c.Machines = &MachinesClient{client: c}
	c.Profiles = &ProfilesClient{client: p}
	c.Teams = &TeamsClient{client: p}
	c.Memberships = &MembershipsClient{client: p}
	c.Invites = &InvitesClient{client: c}
	c.Keypairs = &KeypairsClient{client: c}
	c.Session = &SessionClient{client: c}
	c.Projects = &ProjectsClient{client: p}
	c.Services = &ServicesClient{client: p}
	c.Environments = &EnvironmentsClient{client: p}
	c.Credentials = &CredentialsClient{client: c}
	c.Policies = &PoliciesClient{client: p}
	c.Worklog = &WorklogClient{client: c}
	c.Version = &VersionClient{client: c}

	return c
}

// NewRequest constructs a new http.Request, with a body containing the json
// representation of body, if provided.
func (c *Client) NewRequest(method, path string, query *url.Values, body interface{}, proxied bool) (*http.Request, string, error) {
	requestID := uuid.NewV4().String()

	b := &bytes.Buffer{}
	if body != nil {
		enc := json.NewEncoder(b)
		err := enc.Encode(body)
		if err != nil {
			return nil, requestID, err
		}
	}

	if query == nil {
		query = &url.Values{}
	}

	version := "v1"
	if proxied {
		version = "proxy"
	}

	fullPath := "http://localhost/" + version + path
	if q := query.Encode(); q != "" {
		fullPath += "?" + q
	}

	req, err := http.NewRequest(method, fullPath, b)
	if err != nil {
		return nil, requestID, err
	}

	req.Header.Set("Host", "localhost")
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Content-type", "application/json")

	return req, requestID, nil
}

// Do executes an http.Request, populating v with the JSON response
// on success.
//
// If the request errors with a JSON formatted response body, it will be
// unmarshaled into the returned error.
func (c *Client) Do(ctx context.Context, r *http.Request, v interface{}, reqID *string, progress *ProgressFunc) (*http.Response, error) {
	done := make(chan bool)
	if progress != nil {
		version := "v1"
		req, err := http.NewRequest("GET", "http://localhost/"+version+"/observe", nil)
		if err != nil {
			return nil, err
		}
		stream, err := eventsource.SubscribeWith("", c.client, req)
		if err != nil {
			return nil, err
		}

		output := *progress
		go func() {
			for {
				select {
				case <-done:
					return
				case ev := <-stream.Events:
					data := ev.Data()
					raw := []byte(data)

					event := Event{}
					event.MessageType = "message"
					err = json.Unmarshal(raw, &event)
					if err != nil {
						output(nil, err)
						go func() {
							<-done
						}()
						return
					}
					if event.ID == *reqID {
						output(&event, nil)
					}
				case err := <-stream.Errors:
					output(nil, err)
				}
			}
		}()
	}

	resp, err := c.client.Do(r)
	if progress != nil {
		done <- true
	}
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
			return errors.New("malformed error response from daemon")
		}

		return apitypes.FormatError(rErr)
	}

	return errors.New("error from daemon. Check status code")
}
