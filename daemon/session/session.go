package session

import "errors"
import "sync"
import "fmt"

type memorySession struct {
	token      string
	passphrase string
	mutex      *sync.Mutex
}

type Session interface {
	SetToken(string) error
	SetPassphrase(string) error
	GetToken() string
	GetPassphrase() string
	HasToken() bool
	HasPassphrase() bool
	Logout()
	String() string
}

func NewSession() Session {
	return &memorySession{token: "", passphrase: "", mutex: &sync.Mutex{}}
}

func (s *memorySession) SetToken(token string) error {

	if len(token) == 0 {
		return errors.New("Token must not be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.token = token
	return nil
}

func (s *memorySession) SetPassphrase(passphrase string) error {
	if len(passphrase) == 0 {
		return errors.New("Passphrase must not be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.passphrase = passphrase
	return nil
}

func (s *memorySession) GetToken() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.token
}

func (s *memorySession) GetPassphrase() string {
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
