package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
)

var errBadResponse = errors.New("bad error response body received from server")

// RoundTripper is the interface used to construct and send requests to
// the torus registry.
type RoundTripper interface {
	NewRequest(method, path string, query *url.Values, body interface{}) (*http.Request, error)
	Do(ctx context.Context, r *http.Request, v interface{}) (*http.Response, error)
}

// DefaultRoundTripper is a default implementation of the RoundTripper
// interface. It is shared and extended by the registry and api clients.
type DefaultRoundTripper struct {
	Client *http.Client
	Host   string
}

// NewRequest constructs a new http.Request, with a body containing the json
// representation of body, if provided.
func (rt *DefaultRoundTripper) NewRequest(method, path string,
	query *url.Values, body interface{}) (*http.Request, error) {

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

	if q := query.Encode(); q != "" {
		path += "?" + q
	}

	req, err := http.NewRequest(method, rt.Host+path, b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", rt.Host)

	if body != nil {
		req.Header.Set("Content-type", "application/json")
	}

	return req, nil
}

// Do executes an http.Request, populating v with the JSON response
// on success.
//
// If the request errors with a JSON formatted response body, it will be
// unmarshaled into the returned error.
func (rt *DefaultRoundTripper) Do(ctx context.Context, r *http.Request,
	v interface{}) (*http.Response, error) {

	resp, err := rt.Client.Do(r)
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

// checkReponseCode checks if the response from the server is an error,
// and if so, attempts to marshal the response into the error type.
func checkResponseCode(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	rErr := &apitypes.Error{StatusCode: r.StatusCode}
	if r.ContentLength != 0 {
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(rErr)
		if err != nil {
			return errBadResponse
		}

		return rErr
	}

	return fmt.Errorf("unknown error response from server with status code %d",
		r.StatusCode)
}
