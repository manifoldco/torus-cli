// Package envelope defines the generic encapsulating format for Arigato
// objects.
package envelope

import (
	"encoding/json"
	"fmt"

	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
)

// Envelope is the interface implemented by objects that encapsulate 'true'
// Arigato objects.
type Envelope interface {
	GetID() *identity.ID // avoid field collision
}

// Signed is the generic format for encapsulating signed immutable
// request/response objects to/from arigato.
type Signed struct {
	ID        *identity.ID          `json:"id"`
	Version   uint8                 `json:"version"`
	Body      identity.Identifiable `json:"body"`
	Signature primitive.Signature   `json:"sig"`
}

// GetID returns the ID of the object encapsulated in this envelope.
func (e *Signed) GetID() *identity.ID {
	return e.ID
}

// UnmarshalJSON implements the json.Unmarshaler interface for Signed
// envelopes.
func (e *Signed) UnmarshalJSON(b []byte) error {
	o, body, err := envelopeUnmarshal(b)
	if err != nil {
		return err
	}

	e.ID = o.ID
	e.Version = o.Version
	e.Signature = o.Signature
	e.Body = body

	return nil
}

// Unsigned is the generic format for encapsulating unsigned mutable
// request/response objects to/from arigato.
type Unsigned struct {
	ID      *identity.ID          `json:"id"`
	Version uint8                 `json:"version"`
	Body    identity.Identifiable `json:"body"`
}

// GetID returns the ID of the object encapsulated in this envelope.
func (e *Unsigned) GetID() *identity.ID {
	return e.ID
}

// UnmarshalJSON implements the json.Unmarshaler interface for Unsigned
// envelopes.
func (e *Unsigned) UnmarshalJSON(b []byte) error {
	o, body, err := envelopeUnmarshal(b)
	if err != nil {
		return err
	}

	e.ID = o.ID
	e.Version = o.Version
	e.Body = body

	return nil
}

// Shared unmarshaling for signed and unsigned Envelopes
func envelopeUnmarshal(b []byte) (*outEnvelope, identity.Identifiable, error) {
	o := outEnvelope{}
	err := json.Unmarshal(b, &o)
	if err != nil {
		return nil, nil, err
	}

	var body identity.Identifiable

	t := o.ID.Type()
	switch t {
	case 0x01:
		body = &primitive.User{}
	case 0x06:
		body = &primitive.PublicKey{}
	case 0x07:
		body = &primitive.PrivateKey{}
	case 0x08:
		body = &primitive.Claim{}
	case 0x09:
		body = &primitive.Keyring{}
	case 0x0a:
		body = &primitive.KeyringMember{}
	case 0x0b:
		body = &primitive.Credential{}
	case 0x0d:
		body = &primitive.Org{}
	default:
		return nil, nil, fmt.Errorf("Unknown primitive type id: %d", t)
	}

	err = json.Unmarshal(o.Body, body)

	return &o, body, err
}

type outEnvelope struct {
	ID        *identity.ID        `json:"id"`
	Version   uint8               `json:"version"`
	Body      json.RawMessage     `json:"body"`
	Signature primitive.Signature `json:"sig"`
}
