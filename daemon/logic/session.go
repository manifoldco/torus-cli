package logic

import (
	"context"
	"fmt"
	"log"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"

	"github.com/manifoldco/torus-cli/daemon/crypto"
)

// Session represents the business logic for creating and managing tokens (and
// their underlying effects on the current session)
type Session struct {
	engine *Engine
}

type updateProfile struct {
	Email     string                    `json:"email,omitempty"`
	Name      string                    `json:"name,omitempty"`
	Password  *primitive.UserPassword   `json:"password,omitempty"`
	Master    *primitive.MasterKey      `json:"master,omitempty"`
	PublicKey *primitive.LoginPublicKey `json:"public_key,omitempty"`
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

	salt, loginToken, err := s.engine.client.Tokens.PostLogin(ctx, creds)
	if err != nil {
		return err
	}

	mechanism := loginToken.Body.Mechanism
	var authToken *envelope.Token
	switch mechanism {
	case primitive.HMACAuth:
		authToken, err = s.attemptEdDSAUpgrade(ctx, loginToken, salt, creds)
	case primitive.EdDSAAuth:
		authToken, err = s.attemptEdDSALogin(ctx, loginToken, salt, creds)
	default:
		err = &apitypes.Error{
			Type: apitypes.InternalServerError,
			Err:  []string{fmt.Sprintf("unrecognized auth mechanism: %s", mechanism)},
		}
	}
	if err != nil {
		return err
	}

	token := []byte(authToken.Body.Token)
	self, err := s.engine.client.Self.Get(ctx, authToken.Body.Token)
	if err != nil {
		return err
	}

	s.engine.db.Set(self.Identity)
	if self.Type == apitypes.UserSession {
		s.engine.db.Set(self.Auth)
	}

	return s.engine.session.Set(self.Type, self.Identity, self.Auth, creds.Passphrase(), token)
}

// Verify attempts to verify the users account using the
func (s *Session) Verify(ctx context.Context, code string) error {
	if s.engine.session.Type() == apitypes.MachineSession {
		return &apitypes.Error{
			Type: apitypes.BadRequestError,
			Err:  []string{"A machine cannot verify it's acccount!"},
		}
	}

	err := s.engine.client.Users.VerifyEmail(ctx, code)
	if err != nil {
		return err
	}

	token := s.engine.session.Token()
	self, err := s.engine.client.Self.Get(ctx, string(token))
	if err != nil {
		return err
	}

	err = s.engine.session.SetIdentity(self.Type, self.Identity, self.Auth)
	if err != nil {
		return err
	}

	return s.engine.session.SetIdentity(self.Type, self.Identity, self.Auth)
}

// UpdateProfile attempts to update the root password used by a user to log
// into Torus which also allows them to access their stored and encrypted
// secrets.
func (s *Session) UpdateProfile(ctx context.Context, newEmail, newName, newPassword string) (envelope.UserInf, error) {
	if s.engine.session.Type() != apitypes.UserSession {
		return nil, &apitypes.Error{
			Type: apitypes.BadRequestError,
			Err:  []string{"You must be a logged in user to change your password!"},
		}
	}

	// Convert the auth portion of the session to a UserInf interface
	// Note: the auth and identity sections for a user are the same
	user, ok := s.engine.session.Self().Auth.(envelope.UserInf)
	if !ok {
		log.Printf("Could not convert to UserInf during update profile")
		return nil, &apitypes.Error{
			Type: apitypes.InternalServerError,
			Err:  []string{"Could not convert to user interface"},
		}
	}

	if user.StructVersion() != 2 {
		return nil, &apitypes.Error{
			Type: apitypes.BadRequestError,
			Err: []string{fmt.Sprintf(
				"User schema must be v2 to perform password change: %d",
				user.StructVersion()),
			},
		}
	}

	payload := &updateProfile{}
	if newEmail != "" {
		payload.Email = newEmail
	}

	if newName != "" {
		payload.Name = newName
	}

	if newPassword != "" {
		pw, master, keypair, err := s.engine.crypto.ChangePassword(ctx, newPassword)
		if err != nil {
			log.Printf("Could not re-encrypt master key: %s", err)
			return nil, &apitypes.Error{
				Type: apitypes.InternalServerError,
				Err:  []string{"Could not re-encrypt master key"},
			}
		}

		payload.Password = pw
		payload.Master = master
		payload.PublicKey = keypair
	}

	updatedUser, err := s.engine.client.Users.Update(ctx, payload)
	if err != nil {
		log.Printf("Could not update password on server due to err: %s", err)
		return nil, err
	}

	s.engine.session.SetIdentity(apitypes.UserSession, updatedUser, updatedUser)
	return updatedUser, nil
}

func (s *Session) attemptEdDSAUpgrade(ctx context.Context, loginToken *envelope.Token,
	salt *base64.Value, creds apitypes.LoginCredential) (*envelope.Token, error) {

	tokenString := loginToken.Body.Token
	hmac, err := crypto.DeriveLoginHMAC(ctx, creds.Passphrase(), salt.String(), tokenString)
	if err != nil {
		return nil, err
	}

	pw := []byte(creds.Passphrase())
	keypair, err := crypto.DeriveLoginKeypair(ctx, pw, salt)
	if err != nil {
		return nil, err
	}

	sig := keypair.Sign([]byte(tokenString))
	return s.engine.client.Tokens.PostUpgradeEdDSAAuth(ctx, tokenString, hmac,
		sig, keypair.PublicKey())
}

func (s *Session) attemptEdDSALogin(ctx context.Context, loginToken *envelope.Token,
	salt *base64.Value, creds apitypes.LoginCredential) (*envelope.Token, error) {

	tokenString := loginToken.Body.Token
	pw := []byte(creds.Passphrase())

	keypair, err := crypto.DeriveLoginKeypair(ctx, pw, salt)
	if err != nil {
		return nil, err
	}

	sig := keypair.Sign([]byte(tokenString))
	return s.engine.client.Tokens.PostEdDSAAuth(ctx, tokenString, sig)
}

// Logout destroys the current session if it exists, otherwise, it returns an
// error that the request could not be completed.
func (s *Session) Logout(ctx context.Context) error {

	if !s.engine.session.HasToken() {
		return &apitypes.Error{
			Type: apitypes.UnauthorizedError,
			Err:  []string{"You must be logged in, to logout!"},
		}
	}

	tok := s.engine.session.Token()
	err := s.engine.client.Tokens.Delete(ctx, string(tok[:]))
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
