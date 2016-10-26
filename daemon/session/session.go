// Package session provides in-memory storage of secure session details.
package session

import (
	"errors"
	"fmt"
	"sync"

	"github.com/manifoldco/torus-cli/identity"
)

type memorySession struct {
	id *identity.ID

	// sensitive values
	token      string
	passphrase string
	mutex      *sync.Mutex
}

// Session is the interface for access to secure session details.
type Session interface {
	Set(*identity.ID, string, string) error
	ID() *identity.ID
	Token() string
	Passphrase() string
	HasToken() bool
	HasPassphrase() bool
	Logout()
	String() string
}

// NewSession returns the default implementation of the Session interface.
func NewSession() Session {
	return &memorySession{mutex: &sync.Mutex{}}
}

// Set atomically sets all relevant session details.
//
// It returns an error if any values are empty.
func (s *memorySession) Set(id *identity.ID, passphrase, token string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(id) == 0 {
		return errors.New("ID must not be empty")
	}

	if len(passphrase) == 0 {
		return errors.New("Passphrase must not be empty")
	}

	if len(token) == 0 {
		return errors.New("Token must not be empty")
	}

	s.id = id
	s.passphrase = passphrase
	s.token = token

	return nil
}

func (s *memorySession) ID() *identity.ID {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.id
}

// Token returns the auth token stored in this session.
func (s *memorySession) Token() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.token
}

// Passphrase returns the user's passphrase.
func (s *memorySession) Passphrase() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.passphrase
}

func (s *memorySession) HasToken() bool {
	return (len(s.token) > 0)
}

func (s *memorySession) HasPassphrase() bool {
	return (len(s.passphrase) > 0)
}

// Logout clears all details from the session.
func (s *memorySession) Logout() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id = nil
	s.token = ""
	s.passphrase = ""
}

// String implements the fmt.Stringer interface.
func (s *memorySession) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return fmt.Sprintf(
		"memorySession{token:%t,passphrase:%t}", s.HasToken(), s.HasPassphrase())
}
