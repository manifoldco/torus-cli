// Package envelope defines the generic encapsulating format for torus
// objects.
package envelope

import (
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// Envelope is the interface implemented by objects that encapsulate 'true'
// torus objects.
type Envelope interface {
	GetID() *identity.ID // avoid field collision
}

// Signed is the generic format for encapsulating signed immutable
// request/response objects to/from torus.
type Signed struct {
	ID        *identity.ID        `json:"id"`
	Version   uint8               `json:"version"`
	Body      identity.Immutable  `json:"body"`
	Signature primitive.Signature `json:"sig"`
}

// GetID returns the ID of the object encapsulated in this envelope.
func (e *Signed) GetID() *identity.ID {
	return e.ID
}

// Unsigned is the generic format for encapsulating unsigned mutable
// request/response objects to/from torus.
type Unsigned struct {
	ID      *identity.ID     `json:"id"`
	Version uint8            `json:"version"`
	Body    identity.Mutable `json:"body"`
}

// GetID returns the ID of the object encapsulated in this envelope.
func (e *Unsigned) GetID() *identity.ID {
	return e.ID
}
