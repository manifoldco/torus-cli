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

const daemonAPIVersion = "v1"

// RoundTripper is the interface used to construct and send requests to
// the torus registry.
type RoundTripper interface {
	NewRequest(method, path string, query *url.Values, body interface{}) (*http.Request, error)
	Do(ctx context.Context, r *http.Request, v interface{}) (*http.Response, error)
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
	OrgInvites   *OrgInvitesClient
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

	c.Orgs = &OrgsClient{client: c}
	c.Users = newUsersClient(c)
	c.Machines = newMachinesClient(c)
	c.Profiles = &ProfilesClient{client: c}
	c.Teams = &TeamsClient{client: c}
	c.Memberships = &MembershipsClient{client: c}
	c.OrgInvites = newOrgInvitesClient(c)
	c.Keypairs = newKeypairsClient(c)
	c.Session = &SessionClient{client: c}
	c.Projects = &ProjectsClient{client: c}
	c.Services = &ServicesClient{client: c}
	c.Environments = &EnvironmentsClient{client: c}
	c.Credentials = &CredentialsClient{client: c}
	c.Policies = &PoliciesClient{client: c}
	c.Worklog = &WorklogClient{client: c}
	c.Version = newVersionClient(c)

	return c
}

// NewDaemonRequest constructs a new http.Request, with a body containing the json
// representation of body, if provided. Daemon requests are handled directly
// by the torus daemon.
func (c *Client) NewDaemonRequest(method, path string, query *url.Values, body interface{}) (*http.Request, string, error) {
	return c.newRequest(method, daemonAPIVersion, path, query, body)
}

// NewRequest constructs a new http.Request, with a body containing the json
// representation of body, if provided.
func (c *Client) NewRequest(method string, path string, query *url.Values, body interface{}) (*http.Request, error) {
	req, _, err := c.newRequest(method, "proxy", path, query, body)
	return req, err
}

func (c *Client) newRequest(method, prefix, path string, query *url.Values, body interface{}) (*http.Request, string, error) {
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

	fullPath := "http://localhost/" + prefix + path
	if q := query.Encode(); q != "" {
		fullPath += "?" + q
	}

	req, err := http.NewRequest(method, fullPath, b)
	if err != nil {
		return nil, requestID, err
	}

	req.Header.Set("Host", "localhost")
	req.Header.Set("X-Request-ID", requestID)

	if body != nil {
		req.Header.Set("Content-type", "application/json")
	}

	return req, requestID, nil
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

// DoWithProgress executes the HTTP request like Do, in addition to
// connecting the provided ProgressFunc to any server-sent event progress
// messages.
func (c *Client) DoWithProgress(ctx context.Context, r *http.Request, v interface{}, reqID string, progress ProgressFunc) (*http.Response, error) {
	done := make(chan bool)
	req, _, err := c.newRequest("GET", daemonAPIVersion, "/observe", nil, nil)
	if err != nil {
		return nil, err
	}
	stream, err := eventsource.SubscribeWith("", c.client, req)
	if err != nil {
		return nil, err
	}

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
					progress(nil, err)
					go func() {
						<-done
					}()
					return
				}
				if event.ID == reqID {
					progress(&event, nil)
				}
			case err := <-stream.Errors:
				progress(nil, err)
			}
		}
	}()

	defer func() { done <- true }()

	return c.Do(ctx, r, v)
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
