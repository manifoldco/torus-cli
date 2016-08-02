// Package registry provides access to the Arigato registry REST API.
package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arigatomachine/cli/daemon/session"
)

// Error represents API errors from the registry service.
type Error struct {
	StatusCode int

	Type string   `json:"type"`
	Err  []string `json:"error"`
}

// Error implements the error interface for registry errors.
func (e *Error) Error() string {
	return e.Type
}

// Client exposes the registry REST API.
type Client struct {
	client *http.Client
	prefix string
	sess   session.Session

	KeyPairs *KeyPairs
	Tokens   *Tokens
	Users    *Users
}

// NewClient returns a new Client.
func NewClient(prefix string, sess session.Session, t *http.Transport) *Client {
	c := &Client{
		client: &http.Client{Transport: t},
		prefix: prefix,
		sess:   sess,
	}

	c.KeyPairs = &KeyPairs{client: c}
	c.Tokens = &Tokens{client: c}
	c.Users = &Users{client: c}

	return c
}

// NewRequest constructs a new http.Request, with a body containing the json
// representation of body, if provided.
func (c *Client) NewRequest(method, path string, body interface{}) (
	*http.Request, error) {

	return c.NewTokenRequest(c.sess.Token(), method, path, body)
}

// NewTokenRequest constructs a new http.Request, with a body containing the
// json representation of body, if provided.
//
// The request will be authorized with the provided token.
func (c *Client) NewTokenRequest(token, method, path string, body interface{}) (
	*http.Request, error) {

	b := &bytes.Buffer{}
	if body != nil {
		enc := json.NewEncoder(b)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, c.prefix+path, b)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "bearer "+token)
	}
	req.Header.Set("Content-type", "application/json")

	return req, nil
}

// Do executes an http.Request, populating v with the JSON response
// on success.
//
// If the request errors with a JSON formatted response body, it will be
// unmarshaled into the returned error.
func (c *Client) Do(r *http.Request, v interface{}) (*http.Response, error) {
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

	rErr := &Error{StatusCode: r.StatusCode}
	if r.ContentLength != 0 {
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(rErr)
		if err != nil {
			return errors.New("Malformed error response from registry")
		}

		return rErr
	}

	return errors.New("Error from registry. Check status code")
}
