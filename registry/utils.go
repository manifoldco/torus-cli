package registry

import (
	"context"
	"log"
	"net/http"
	"net/url"
)

func replaceAuthToken(req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Del("Authorization")
	}
}

func tokenRoundTrip(ctx context.Context, rd RequestDoer, token, method,
	path string, query *url.Values, body, response interface{}) error {

	req, err := rd.NewRequest(method, path, query, body)
	replaceAuthToken(req, token)
	if err != nil {
		log.Printf("Error building request: %s", err)
		return err
	}

	_, err = rd.Do(ctx, req, response)
	if err != nil {
		log.Printf("Error making request: %s", err)
	}

	return err
}
