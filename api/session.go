package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

var errUnknownSessionType = errors.New("Unknown session type")

// Session represents the current logged in entities identity and
// authentication objects which could be a Machine or User
type Session struct {
	sessionType apitypes.SessionType
	identity    envelope.Envelope
	auth        envelope.Envelope
}

// Type returns the type of the session (machine or user)
func (s *Session) Type() apitypes.SessionType {
	return s.sessionType
}

// ID returns the identity ID (e.g. user or machine id)
func (s *Session) ID() *identity.ID {
	return s.identity.GetID()
}

// AuthID returns the auth id (e.g. user or machine token id)
func (s *Session) AuthID() *identity.ID {
	return s.auth.GetID()
}

// Username returns the username or machine name depending on the session type
func (s *Session) Username() string {
	if s.sessionType == apitypes.MachineSession {
		return s.identity.(*envelope.Machine).Body.Name
	}

	return s.identity.(envelope.UserInf).Username()
}

// Name returns the fullname of the user or the machine name
func (s *Session) Name() string {
	if s.sessionType == apitypes.MachineSession {
		return s.identity.(*envelope.Machine).Body.Name
	}

	return s.identity.(envelope.UserInf).Name()
}

// Email returns none for a machine or the users email address
func (s *Session) Email() string {
	if s.sessionType == apitypes.MachineSession {
		return "none"
	}

	return s.identity.(envelope.UserInf).Email()
}

// NewSession returns a new session constructed from the payload of the current
// identity as returned from the Daemon
func NewSession(resp *apitypes.Self) (*Session, error) {
	switch resp.Type {
	case apitypes.UserSession:
		if _, ok := resp.Identity.(envelope.UserInf); !ok {
			return nil, errors.New("Identity must be a user object")
		}
		if _, ok := resp.Auth.(envelope.UserInf); !ok {
			return nil, errors.New("Auth must be a user object")
		}
	case apitypes.MachineSession:
		if _, ok := resp.Identity.(*envelope.Machine); !ok {
			return nil, errors.New("Identity must be a machine object")
		}
		if _, ok := resp.Auth.(*envelope.MachineToken); !ok {
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
	client *apiRoundTripper
}

type rawSelf struct {
	*apitypes.Self

	// Shadow over the pure self value
	Identity json.RawMessage `json:"identity"`
	Auth     json.RawMessage `json:"auth"`
}

// Who returns the Session object for the current authenticated user or machine
func (s *SessionClient) Who(ctx context.Context) (*Session, error) {
	raw := &rawSelf{}
	err := s.client.DaemonRoundTrip(ctx, "GET", "/self", nil, nil, raw, nil)
	if err != nil {
		return nil, err
	}

	self := raw.Self
	switch raw.Type {
	case apitypes.UserSession:
		self.Identity = &envelope.User{}
		self.Auth = &envelope.User{}
	case apitypes.MachineSession:
		self.Identity = &envelope.Machine{}
		self.Auth = &envelope.MachineToken{}
	default:
		return nil, errUnknownSessionType
	}

	err = json.Unmarshal(raw.Identity, self.Identity)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw.Auth, self.Auth)
	if err != nil {
		return nil, err
	}

	return NewSession(self)
}

// Get returns the status of the user's session.
func (s *SessionClient) Get(ctx context.Context) (*apitypes.SessionStatus, error) {
	resp := &apitypes.SessionStatus{}
	err := s.client.DaemonRoundTrip(ctx, "GET", "/session", nil, nil, resp, nil)
	return resp, err
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

func performLogin(ctx context.Context, s *SessionClient, loginType apitypes.SessionType, rawLogin json.RawMessage) error {
	wrapper := apitypes.Login{
		Type:        loginType,
		Credentials: rawLogin,
	}

	return s.client.DaemonRoundTrip(ctx, "POST", "/login", nil, &wrapper, nil, nil)
}

// Logout logs the user out of their session
func (s *SessionClient) Logout(ctx context.Context) error {
	return s.client.DaemonRoundTrip(ctx, "POST", "/logout", nil, nil, nil, nil)
}
