// Package session provides in-memory storage of secure session details.
package session

import (
	"errors"
	"fmt"
	"sync"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/daemon/crypto/secure"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

const notLoggedInError = "Please login to perform that command"

type session struct {
	mutex       *sync.Mutex
	sessionType apitypes.SessionType

	// XXX: These should be scoped down to an interface that can do auth stuff.
	identity envelope.Envelope
	auth     envelope.Envelope

	// sensitive values and the guard
	guard      *secure.Guard
	token      *secure.Secret
	passphrase *secure.Secret
}

// Session is the interface for access to secure session details.
type Session interface {
	Type() apitypes.SessionType
	Set(apitypes.SessionType, envelope.Envelope, envelope.Envelope, []byte, []byte) error
	SetIdentity(apitypes.SessionType, envelope.Envelope, envelope.Envelope) error
	ID() *identity.ID
	AuthID() *identity.ID
	Token() []byte
	Passphrase() []byte
	MasterKey() (*base64.Value, error)
	HasToken() bool
	HasPassphrase() bool
	Logout() error
	String() string
	Self() *apitypes.Self
}

// NewSession returns the default implementation of the Session interface
// for a user or machine depending on the passed type.
func NewSession(guard *secure.Guard) Session {
	return &session{
		mutex:       &sync.Mutex{},
		sessionType: apitypes.NotLoggedIn,
		guard:       guard,
	}
}

// Type returns the type of identity this session represents (e.g. user or
// machine)
func (s *session) Type() apitypes.SessionType {
	return s.sessionType
}

// ID returns the ID representing the Identity providing object (e.g. user or
// machine)
func (s *session) ID() *identity.ID {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.identity == nil {
		return nil
	}

	return s.identity.GetID()
}

// AuthID returns the ID representing the object used for authorization (e.g.
// user or machine token).
func (s *session) AuthID() *identity.ID {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.auth == nil {
		return nil
	}

	return s.auth.GetID()
}

// Token returns the auth token stored in this session.
func (s *session) Token() []byte {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.token.Buffer()
}

// Passphrase returns the user's passphrase.
func (s *session) Passphrase() []byte {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.passphrase.Buffer()
}

func (s *session) HasToken() bool {
	return s.token != nil
}

func (s *session) HasPassphrase() bool {
	return s.passphrase != nil
}

// String implements the fmt.Stringer interface.
func (s *session) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return fmt.Sprintf("Session{type:%s,token:%t,passphrase:%t}",
		s.Type(), s.HasToken(), s.HasPassphrase())
}

func checkSessionType(sessionType apitypes.SessionType, identity, auth envelope.Envelope) error {
	if identity == nil || auth == nil {
		return errors.New("identity and auth cannot be null")
	}

	switch sessionType {
	case apitypes.UserSession:
		if _, ok := identity.(envelope.UserInf); !ok {
			return errors.New("Identity must be a user object")
		}

		if _, ok := auth.(envelope.UserInf); !ok {
			return errors.New("Auth must be a user object")
		}
	case apitypes.MachineSession:
		if _, ok := identity.(*envelope.Machine); !ok {
			return errors.New("Identity must be machine object")
		}

		if _, ok := auth.(*envelope.MachineToken); !ok {
			return errors.New("Auth must be a machine token object")
		}
	default:
		panic("did not recognize session type")
	}
	return nil
}

// Set atomically sets all relevant session details.
//
// It returns an error if any values are empty.
func (s *session) Set(sessionType apitypes.SessionType, identity, auth envelope.Envelope,
	passphrase []byte, token []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := checkSessionType(sessionType, identity, auth)
	if err != nil {
		return err
	}

	if len(passphrase) == 0 {
		return errors.New("Passphrase must not be empty")
	}

	if len(token) == 0 && sessionType == apitypes.UserSession {
		return errors.New("Token must not be empty")
	}

	s.sessionType = sessionType
	s.passphrase, err = s.guard.Secret(passphrase)
	if err != nil {
		return err
	}

	s.token, err = s.guard.Secret(token)
	if err != nil {
		return err
	}

	s.identity = identity
	s.auth = auth

	return nil
}

// SetIdentity updates the session identity
func (s *session) SetIdentity(sessionType apitypes.SessionType, identity, auth envelope.Envelope) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := checkSessionType(sessionType, identity, auth)
	if err != nil {
		return err
	}

	s.identity = identity
	s.auth = auth

	return nil
}

func createNotLoggedInError() error {
	return &apitypes.Error{
		Type: apitypes.UnauthorizedError,
		Err:  []string{notLoggedInError},
	}
}

// Returns the base64 representation of the identities encrypted master key
func (s *session) MasterKey() (*base64.Value, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.Type() == apitypes.NotLoggedIn {
		return nil, createNotLoggedInError()
	}

	if s.Type() == apitypes.UserSession {
		return s.auth.(envelope.UserInf).Master().Value, nil
	}

	return s.auth.(*envelope.MachineToken).Body.Master.Value, nil
}

// Self returns the Self apitype which represents the current sessions state
func (s *session) Self() *apitypes.Self {
	return &apitypes.Self{
		Type:     s.Type(),
		Identity: s.identity,
		Auth:     s.auth,
	}
}

// Logout resets all values to the logged out state
func (s *session) Logout() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.Type() == apitypes.NotLoggedIn {
		return createNotLoggedInError()
	}

	s.sessionType = apitypes.NotLoggedIn
	s.identity = nil
	s.auth = nil

	s.token.Destroy()
	s.passphrase.Destroy()
	s.token = nil
	s.passphrase = nil

	return nil
}
