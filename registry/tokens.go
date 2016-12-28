package registry

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
)

// token types that can be requested from the registry
const (
	tokenTypeLogin = "login"
	tokenTypeAuth  = "auth"
)

type loginTokenUserRequest struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}

type loginTokenMachineRequest struct {
	Type    string `json:"type"`
	TokenID string `json:"machine_token_id"`
}

type loginTokenResponse struct {
	Salt  *base64.Value `json:"salt"`
	Token string        `json:"login_token"`
}

type authTokenHMACRequest struct {
	Type      string `json:"type"`
	TokenHMAC string `json:"login_token_hmac"`
}

type authTokenPDPKARequest struct {
	Type     string        `json:"type"`
	TokenSig *base64.Value `json:"login_token_sig"`
}

type authTokenResponse struct {
	Token string `json:"auth_token"`
}

// TokensClient represents the registry '/tokens' endpoints, used for session
// management.
//
// Logging in is a two step process. We must first request a login token.
// This token is then HMAC'd and returned to the server, exchanging it for
// an auth token, which is used for all other operations.
type TokensClient struct {
	client RequestDoer
}

// PostLogin requests a login token from the registry for the provided email
// address.
func (t *TokensClient) PostLogin(ctx context.Context, creds apitypes.LoginCredential) (*base64.Value, string, error) {
	salt := loginTokenResponse{}

	var body interface{}
	switch creds.Type() {
	case apitypes.UserSession:
		body = &loginTokenUserRequest{
			Type:  tokenTypeLogin,
			Email: creds.Identifier(),
		}
	case apitypes.MachineSession:
		body = &loginTokenMachineRequest{
			Type:    tokenTypeLogin,
			TokenID: creds.Identifier(),
		}
	}

	req, err := t.client.NewRequest("POST", "/tokens", nil, body)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return salt.Salt, salt.Token, err
	}

	resp, err := t.client.Do(ctx, req, &salt)
	if err != nil && resp != nil && resp.StatusCode != 201 {
		log.Printf("Failed to get login token from server: %s", err)
	} else if err != nil {
		log.Printf("Error making api request: %s", err)
	}

	return salt.Salt, salt.Token, err
}

// PostAuth requests an auth token from the registry for the provided login
// token value, and it's HMAC.
func (t *TokensClient) PostAuth(ctx context.Context, token, hmac string) (string, error) {
	authReq := authTokenHMACRequest{Type: tokenTypeAuth, TokenHMAC: hmac}
	return t.postAuthWorker(ctx, token, &authReq)
}

// PostPDPKAuth requests an auth token from the registry for the provided login
// token value, and it's signature.
func (t *TokensClient) PostPDPKAuth(ctx context.Context, token string, sig *base64.Value) (string, error) {
	authReq := authTokenPDPKARequest{Type: tokenTypeAuth, TokenSig: sig}
	return t.postAuthWorker(ctx, token, &authReq)
}

func (t *TokensClient) postAuthWorker(ctx context.Context, token string, authReq interface{}) (string, error) {
	auth := authTokenResponse{}
	err := tokenRoundTrip(ctx, t.client, token, "POST", "/tokens", nil, authReq, &auth)
	return auth.Token, err
}

// Delete deletes the token with the provided value from the registry. This
// effectively logs a user out.
func (t *TokensClient) Delete(ctx context.Context, token string) error {
	return tokenRoundTrip(ctx, t.client, token, "DELETE", "/tokens/"+token, nil, nil, nil)
}
