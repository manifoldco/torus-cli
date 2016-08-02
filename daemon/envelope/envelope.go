package envelope

import (
	"encoding/json"
	"fmt"

	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
)

type Envelope interface {
	GetID() *identity.ID // avoid field collision
}

// Signed is the generic format for encapsulating signed
// request/response objects to/from arigato.
type Signed struct {
	ID        *identity.ID        `json:"id"`
	Version   uint8               `json:"version"`
	Body      identity.AgObject   `json:"body"`
	Signature primitive.Signature `json:"sig"`
}

func (e *Signed) GetID() *identity.ID {
	return e.ID
}

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

type Unsigned struct {
	ID      *identity.ID      `json:"id"`
	Version uint8             `json:"version"`
	Body    identity.AgObject `json:"body"`
}

func (e *Unsigned) GetID() *identity.ID {
	return e.ID
}

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
func envelopeUnmarshal(b []byte) (*outEnvelope, identity.AgObject, error) {
	o := outEnvelope{}
	err := json.Unmarshal(b, &o)
	if err != nil {
		return nil, nil, err
	}

	var body identity.AgObject

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
	default:
		return nil, nil, fmt.Errorf("Unknown type: %d", t)
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
