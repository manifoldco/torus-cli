// Package secure is a wrapper package around memguard for managing data which
// must not be swapped to disk or easily read via a memory scanner.
package secure

import (
	"errors"
	"sync"

	"github.com/awnumar/memguard"
)

// ErrSecretNotFound occurs when a secret could not be found
var ErrSecretNotFound = errors.New("Could not find secret to remove it")

var lock *sync.Mutex
var current *Guard

func init() {
	lock = &sync.Mutex{}
	memguard.DisableUnixCoreDumps()
}

// Guard is a controller of guarded memory which contains many different
// secrets. Only one guard can exist at a time.
type Guard struct {
	secrets []*Secret
	lock    *sync.Mutex
}

// Secret represent a specific guraded piece of memory
type Secret struct {
	guard  *Guard
	buffer *memguard.LockedBuffer
}

// NewGuard returns a Guard instance or errors if a guard already exists.
func NewGuard() *Guard {
	lock.Lock()
	defer lock.Unlock()

	if current != nil {
		return current
	}

	current = &Guard{
		secrets: []*Secret{},
		lock:    &sync.Mutex{},
	}

	return current
}

func (g *Guard) remove(s *Secret) error {
	if s == nil {
		panic("received a nil secret; should not be possible")
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	placement := -1
	for i, v := range g.secrets {
		if *v == *s {
			placement = i
			break
		}
	}

	if placement == -1 {
		return ErrSecretNotFound
	}

	g.secrets = append(g.secrets[:placement], g.secrets[placement+1:]...)

	return nil
}

// Destroy removes all secrets and cleans up everything up appropriately.
func (g *Guard) Destroy() {
	lock.Lock()
	defer lock.Unlock()

	for _, s := range g.secrets {
		s.Destroy()
	}

	current = nil
}

// Secret returns a new secret for the given value, the given byte slice will
// be wiped *after* the secret has been moved into secure memory.
func (g *Guard) Secret(v []byte) (*Secret, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	b, err := memguard.NewMutableFromBytes(v)
	if err != nil {
		return nil, err
	}

	s := Secret{
		buffer: b,
		guard:  g,
	}

	g.secrets = append(g.secrets, &s)
	return &s, nil
}

// Random returns a new secret of the given length in bytes.
func (g *Guard) Random(size int) (*Secret, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	b, err := memguard.NewImmutableRandom(size)
	if err != nil {
		return nil, err
	}

	s := Secret{
		buffer: b,
		guard:  g,
	}

	g.secrets = append(g.secrets, &s)
	return &s, nil
}

// Buffer returns an array of bytes referencing the underlying secure memory.
func (s *Secret) Buffer() []byte {
	return s.buffer.Buffer()
}

// Destroy properly dispenses of the underlying secret stored in secure memory.
func (s *Secret) Destroy() {
	defer s.buffer.Destroy()
	s.guard.remove(s)
}
