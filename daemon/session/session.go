package session

import "errors"
import "sync"
import "fmt"

type memorySession struct {
	id string
	// sensitive values
	token      string
	passphrase string
	mutex      *sync.Mutex
}

type Session interface {
	Set(string, string, string) error
	ID() string
	Token() string
	Passphrase() string
	HasToken() bool
	HasPassphrase() bool
	Logout()
	String() string
}

func NewSession() Session {
	return &memorySession{mutex: &sync.Mutex{}}
}

func (s *memorySession) Set(id, passphrase, token string) error {
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

func (s *memorySession) ID() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.id
}

func (s *memorySession) Token() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.token
}

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

func (s *memorySession) Logout() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.token = ""
	s.passphrase = ""
}

func (s *memorySession) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return fmt.Sprintf(
		"memorySession{token:%t,passphrase:%t}", s.HasToken(), s.HasPassphrase())
}
