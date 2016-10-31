package api

import (
	"context"
	"errors"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

type session struct {
	sessionType string
	identity    *envelope.Unsigned
	auth        *envelope.Unsigned
}

// Session represents the current logged in entities identity and
// authentication objects which could be a Machine or User
type Session interface {
	Type() string
	ID() *identity.ID
	AuthID() *identity.ID
	Username() string
	Name() string
	Email() string
}

func (s *session) Type() string {
	return s.sessionType
}

func (s *session) ID() *identity.ID {
	return s.identity.ID
}

func (s *session) AuthID() *identity.ID {
	return s.auth.ID
}

func (s *session) Username() string {
	if s.sessionType == apitypes.MachineSession {
		return s.identity.Body.(*primitive.Machine).Name
	}

	return s.identity.Body.(*primitive.User).Username
}

func (s *session) Name() string {
	if s.sessionType == apitypes.MachineSession {
		return s.identity.Body.(*primitive.Machine).Name
	}

	return s.identity.Body.(*primitive.User).Name
}

func (s *session) Email() string {
	if s.sessionType == apitypes.MachineSession {
		return "none"
	}

	return s.identity.Body.(*primitive.User).Email
}

// NewSession returns a new session constructed from the payload of the current
// identity as returned from the Daemon
func NewSession(resp *apitypes.Self) (Session, error) {
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

	return &session{
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
func (s *SessionClient) Who(ctx context.Context) (Session, error) {
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

// Login logs the user in using the provided email and passphrase
func (s *SessionClient) Login(ctx context.Context, email, passphrase string) error {
	login := apitypes.Login{
		Email:      email,
		Passphrase: passphrase,
	}
	req, _, err := s.client.NewRequest("POST", "/login", nil, &login, false)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil

}

// Logout logs the user out of their session
func (s *SessionClient) Logout(ctx context.Context) error {
	req, _, err := s.client.NewRequest("POST", "/logout", nil, nil, false)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
