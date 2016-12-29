package registry

import (
	"context"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"
)

type loginTokenUserRequest struct {
	Type  primitive.TokenType `json:"type"`
	Email string              `json:"email"`
}

type loginTokenMachineRequest struct {
	Type    primitive.TokenType `json:"type"`
	TokenID string              `json:"machine_token_id"`
}

type loginTokenResponse struct {
	Token *envelope.Token `json:"token"`
	Salt  *base64.Value   `json:"salt"`
}

type authTokenEdDSARequest struct {
	Type     primitive.TokenType `json:"type"`
	TokenSig *base64.Value       `json:"login_token_sig"`
}

type authTokenUpgradeEdDSARequest struct {
	Type      primitive.TokenType `json:"type"`
	TokenSig  *base64.Value       `json:"login_token_sig"`
	TokenHMAC string              `json:"login_token_hmac"`
	PublicKey *base64.Value       `json:"login_public_key"`
}

type authTokenResponse struct {
	Token *envelope.Token `json:"token"`
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
func (t *TokensClient) PostLogin(ctx context.Context, creds apitypes.LoginCredential) (
	*base64.Value, *envelope.Token, error) {
	respBody := loginTokenResponse{}

	var body interface{}
	switch creds.Type() {
	case apitypes.UserSession:
		body = &loginTokenUserRequest{
			Type:  primitive.LoginToken,
			Email: creds.Identifier(),
		}
	case apitypes.MachineSession:
		body = &loginTokenMachineRequest{
			Type:    primitive.LoginToken,
			TokenID: creds.Identifier(),
		}
	}

	err := tokenRoundTrip(ctx, t.client, "", "POST", "/tokens", nil, body, &respBody)
	if err != nil {
		return nil, nil, err
	}

	return respBody.Salt, respBody.Token, nil
}

// PostEdDSAAuth requests an auth token from the registry for the provided login
// token value, and it's signature.
func (t *TokensClient) PostEdDSAAuth(ctx context.Context, token string,
	sig *base64.Value) (*envelope.Token, error) {
	authReq := authTokenEdDSARequest{Type: primitive.AuthToken, TokenSig: sig}
	return t.postAuthWorker(ctx, token, &authReq)
}

// PostUpgradeEdDSAAuth requests an auth token from the registry while
// upgrading the user from HMAC based authentication to EdDSA
func (t *TokensClient) PostUpgradeEdDSAAuth(ctx context.Context, token, hmac string,
	sig, publicKey *base64.Value) (*envelope.Token, error) {
	authReq := authTokenUpgradeEdDSARequest{
		Type:      primitive.AuthToken,
		TokenSig:  sig,
		PublicKey: publicKey,
		TokenHMAC: hmac,
	}

	return t.postAuthWorker(ctx, token, &authReq)
}

func (t *TokensClient) postAuthWorker(ctx context.Context, token string,
	authReq interface{}) (*envelope.Token, error) {
	auth := authTokenResponse{}
	err := tokenRoundTrip(ctx, t.client, token, "POST", "/tokens", nil, authReq, &auth)
	if err != nil {
		return nil, err
	}

	return auth.Token, err
}

// Delete deletes the token with the provided value from the registry. This
// effectively logs a user out.
func (t *TokensClient) Delete(ctx context.Context, token string) error {
	return tokenRoundTrip(ctx, t.client, token, "DELETE", "/tokens/"+token, nil, nil, nil)
}
