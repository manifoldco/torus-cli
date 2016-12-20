package logic

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/registry"
	"github.com/manifoldco/torus-cli/daemon/session"
)

// Session represents the business logic for creating and managing tokens (and
// their underlying effects on the current session)
type Session struct {
	engine *Engine
}

// Login attempts to create a valid auth token to authorize http requests made
// against the registry.
func (s *Session) Login(ctx context.Context, creds apitypes.LoginCredential) error {
	if !creds.Valid() {
		return &apitypes.Error{
			Type: apitypes.BadRequestError,
			Err:  []string{"invalid login credentials provided"},
		}
	}

	var authToken string
	var err error
	switch creds.Type() {
	case apitypes.UserSession:
		authToken, err = attemptHMACLogin(ctx, s.engine.client, s.engine.session, creds)
	case apitypes.MachineSession:
		authToken, err = attemptPDPKALogin(ctx, s.engine.client, s.engine.session, creds)
	}
	if err != nil {
		return err
	}

	self, err := s.engine.client.Self.Get(ctx, authToken)
	if err != nil {
		return err
	}

	s.engine.db.Set(self.Identity)
	if self.Type == apitypes.UserSession {
		s.engine.db.Set(self.Auth)
	}

	return s.engine.session.Set(self.Type, self.Identity, self.Auth, creds.Passphrase(), authToken)
}

// Logout destroys the current session if it exists, otherwise, it returns an
// error that the request could not be completed.
func (s *Session) Logout(ctx context.Context) error {
	tok := s.engine.session.Token()

	if tok == "" {
		return &apitypes.Error{
			Type: apitypes.UnauthorizedError,
			Err:  []string{"You must be logged in, to logout!"},
		}
	}

	err := s.engine.client.Tokens.Delete(ctx, tok)
	switch err := err.(type) {
	case *apitypes.Error:
		switch {
		case err.StatusCode >= 500:
			// On a 5XX response, we don't know for sure that the server
			// has successfully removed the auth token. Keep the copy in
			// the daemon, so the user may try again.
			return err
		case err.StatusCode >= 400:
			// A 4XX error indicates either the token isn't found, or we're
			// not allowed to remove it (or the server is a teapot).
			//
			// In any case, the daemon has gotten out of sync with the
			// server. Remove our local copy of the auth token.
			log.Printf("Got 4XX removing auth token. Treating as success")
			logoutErr := s.engine.session.Logout()
			if logoutErr != nil {
				return logoutErr
			}

			return nil
		}
	case nil:
		logoutErr := s.engine.session.Logout()
		if logoutErr != nil {
			return logoutErr
		}

		return nil
	default:
		return err
	}

	return nil
}

func attemptPDPKALogin(ctx context.Context, client *registry.Client, s session.Session, creds apitypes.LoginCredential) (string, error) {
	salt, loginToken, err := client.Tokens.PostLogin(ctx, creds)
	if err != nil {
		return "", err
	}

	pw := base64.NewValue([]byte(creds.Passphrase()))
	keypair, err := crypto.DeriveLoginKeypair(ctx, pw, salt)
	if err != nil {
		return "", err
	}

	loginTokenSig := keypair.Sign([]byte(loginToken))
	return client.Tokens.PostPDPKAuth(ctx, loginToken, loginTokenSig)
}

func attemptHMACLogin(ctx context.Context, client *registry.Client, s session.Session, creds apitypes.LoginCredential) (string, error) {
	salt, loginToken, err := client.Tokens.PostLogin(ctx, creds)
	if err != nil {
		return "", err
	}

	hmac, err := crypto.DeriveLoginHMAC(ctx, creds.Passphrase(), salt.String(), loginToken)
	if err != nil {
		return "", err
	}

	return client.Tokens.PostAuth(ctx, loginToken, hmac)
}
