package registry

import (
	"context"
	"log"
)

// token types that can be requested from the registry
const (
	tokenTypeLogin = "login"
	tokenTypeAuth  = "auth"
)

type loginTokenRequest struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}

type loginTokenResponse struct {
	Salt  string `json:"salt"`
	Token string `json:"login_token"`
}

type authTokenRequest struct {
	Type      string `json:"type"`
	TokenHMAC string `json:"login_token_hmac"`
}

type authTokenResponse struct {
	Token string `json:"auth_token"`
}

// Tokens represents the registry '/tokens' endpoints, used for session
// management.
//
// Logging in is a two step process. We must first request a login token.
// This token is then HMAC'd and returned to the server, exchanging it for
// an auth token, which is used for all other operations.
type Tokens struct {
	client *Client
}

// PostLogin requests a login token from the registry for the provided email
// address.
func (t *Tokens) PostLogin(email string) (string, string, error) {
	salt := loginTokenResponse{}

	req, err := t.client.NewRequest("POST", "/tokens", nil,
		&loginTokenRequest{
			Type:  tokenTypeLogin,
			Email: email,
		})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return salt.Salt, salt.Token, err
	}

	resp, err := t.client.Do(context.TODO(), req, &salt)
	if err != nil && resp != nil && resp.StatusCode != 201 {
		log.Printf("Failed to get login token from server: %s", err)
	} else if err != nil {
		log.Printf("Error making api request: %s", err)
	}

	return salt.Salt, salt.Token, err
}

// PostAuth requests an auth token from the registry for the provided login
// token value, and it's HMAC.
func (t *Tokens) PostAuth(token, hmac string) (string, error) {
	auth := authTokenResponse{}

	req, err := t.client.NewTokenRequest(token, "POST", "/tokens", nil,
		&authTokenRequest{Type: tokenTypeAuth, TokenHMAC: hmac})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return auth.Token, err
	}

	_, err = t.client.Do(context.TODO(), req, &auth)
	if err != nil {
		log.Printf("Error making api request: %s", err)
	}

	return auth.Token, err
}

// Delete deletes the token with the provided value from the registry. This
// effectively logs a user out.
func (t *Tokens) Delete(token string) error {
	req, err := t.client.NewTokenRequest(token, "DELETE", "/tokens/"+token, nil, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return err
	}

	_, err = t.client.Do(context.TODO(), req, nil)
	if err != nil {
		log.Printf("Error making api request: %s", err)
	}

	return err
}
