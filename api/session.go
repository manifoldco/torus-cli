package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// Session represents the current logged in entities identity and
// authentication objects which could be a Machine or User
type Session struct {
	sessionType string
	identity    *envelope.Unsigned
	auth        *envelope.Unsigned
}

// Type returns the type of the session (machine or user)
func (s *Session) Type() string {
	return s.sessionType
}

// ID returns the identity ID (e.g. user or machine id)
func (s *Session) ID() *identity.ID {
	return s.identity.ID
}

// AuthID returns the auth id (e.g. user or machine token id)
func (s *Session) AuthID() *identity.ID {
	return s.auth.ID
}

// Username returns the username or machine name depending on the session type
func (s *Session) Username() string {
	if s.sessionType == apitypes.MachineSession {
		return s.identity.Body.(*primitive.Machine).Name
	}

	return s.identity.Body.(*primitive.User).Username
}

// Name returns the fullname of the user or the machine name
func (s *Session) Name() string {
	if s.sessionType == apitypes.MachineSession {
		return s.identity.Body.(*primitive.Machine).Name
	}

	return s.identity.Body.(*primitive.User).Name
}

// Email returns none for a machine or the users email address
func (s *Session) Email() string {
	if s.sessionType == apitypes.MachineSession {
		return "none"
	}

	return s.identity.Body.(*primitive.User).Email
}

// NewSession returns a new session constructed from the payload of the current
// identity as returned from the Daemon
func NewSession(resp *apitypes.Self) (*Session, error) {
	switch resp.Type {
	case apitypes.UserSession:
		if _, ok := resp.Identity.Body.(*primitive.User); !ok {
			return nil, errors.New("Identity must be a user object")
		}
		if _, ok := resp.Auth.Body.(*primitive.User); !ok {
			return nil, errors.New("Auth must be a user object")
		}
	case apitypes.MachineSession:
		if _, ok := resp.Identity.Body.(*primitive.Machine); !ok {
			return nil, errors.New("Identity must be a machine object")
		}
		if _, ok := resp.Auth.Body.(*primitive.MachineToken); !ok {
			return nil, errors.New("Auth must be a machine token object")
		}

	default:
		return nil, errors.New("did not recognize session type")
	}

	return &Session{
		sessionType: resp.Type,
		identity:    resp.Identity,
		auth:        resp.Auth,
	}, nil
}

// SessionClient provides access to the daemon's user session related endpoints,
// for logging in/out, and checking your session status.
type SessionClient struct {
	client *Client
}

// Who returns the Session object for the current authenticated user or machine
func (s *SessionClient) Who(ctx context.Context) (*Session, error) {
	req, _, err := s.client.NewRequest("GET", "/self", nil, nil, false)
	if err != nil {
		return nil, err
	}

	resp := &apitypes.Self{}
	_, err = s.client.Do(ctx, req, resp, nil, nil)
	if err != nil {
		return nil, err
	}

	return NewSession(resp)
}

// Get returns the status of the user's session.
func (s *SessionClient) Get(ctx context.Context) (*apitypes.SessionStatus, error) {
	req, _, err := s.client.NewRequest("GET", "/session", nil, nil, false)
	if err != nil {
		return nil, err
	}

	resp := &apitypes.SessionStatus{}
	_, err = s.client.Do(ctx, req, resp, nil, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UserLogin logs the user in using the provided email and passphrase
func (s *SessionClient) UserLogin(ctx context.Context, email, passphrase string) error {
	// Package up login credentials for the user
	login := apitypes.UserLogin{
		Email:    email,
		Password: passphrase,
	}

	rawLogin, err := json.Marshal(login)
	if err != nil {
		return err
	}

	return performLogin(ctx, s, "user", rawLogin)
}

// MachineLogin logs the user in using the provided token id and secret
func (s *SessionClient) MachineLogin(ctx context.Context, tokenID, tokenSecret string) error {

	ID, err := identity.DecodeFromString(tokenID)
	if err != nil {
		return err
	}

	secret, err := base64.NewValueFromString(tokenSecret)
	if err != nil {
		return err
	}

	login := apitypes.MachineLogin{
		TokenID: &ID,
		Secret:  secret,
	}

	rawLogin, err := json.Marshal(login)
	if err != nil {
		return err
	}
	return performLogin(ctx, s, "machine", rawLogin)
}

func performLogin(ctx context.Context, s *SessionClient, loginType string, rawLogin json.RawMessage) error {
	wrapper := apitypes.Login{
		Type:        loginType,
		Credentials: rawLogin,
	}

	req, _, err := s.client.NewRequest("POST", "/login", nil, &wrapper, false)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil, nil, nil)
	return err
}

// Logout logs the user out of their session
func (s *SessionClient) Logout(ctx context.Context) error {
	req, _, err := s.client.NewRequest("POST", "/logout", nil, nil, false)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil, nil, nil)
	return err
}
