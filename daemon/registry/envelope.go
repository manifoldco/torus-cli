package registry

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/arigatomachine/cli/daemon/crypto"
)

// Envelope is the generic format for encapsulating request/response objects
// to/from arigato.
type Envelope struct {
	ID        *ID       `json:"id"`
	Version   uint8     `json:"version"`
	Body      AgObject  `json:"body"`
	Signature Signature `json:"sig"`
}

func NewEnvelope(engine *crypto.Engine, body AgObject, sigID *ID,
	sigKP *crypto.SignatureKeyPair) (*Envelope, error) {

	b, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	s, err := engine.Sign(*sigKP,
		append([]byte(strconv.Itoa(body.Version())), b...))
	if err != nil {
		return nil, err
	}

	sv := Base64Value(s)
	sig := Signature{
		PublicKeyID: sigID,
		Algorithm:   crypto.EdDSA,
		Value:       &sv,
	}

	id, err := NewID(body, &sig)
	if err != nil {
		return nil, err
	}

	return &Envelope{
		ID:        &id,
		Version:   1,
		Body:      body,
		Signature: sig,
	}, nil
}

type outEnvelope struct {
	ID        *ID             `json:"id"`
	Version   uint8           `json:"version"`
	Body      json.RawMessage `json:"body"`
	Signature Signature       `json:"sig"`
}

func (e *Envelope) UnmarshalJSON(b []byte) error {
	o := outEnvelope{}
	err := json.Unmarshal(b, &o)
	if err != nil {
		return err
	}

	e.ID = o.ID
	e.Version = o.Version
	e.Signature = o.Signature

	t := o.ID.Type()
	switch t {
	case 0x06:
		e.Body = &PublicKey{}
	case 0x07:
		e.Body = &PrivateKey{}
	case 0x08:
		e.Body = &Claim{}
	default:
		return fmt.Errorf("Unknown type: %s", t)
	}

	return json.Unmarshal(o.Body, e.Body)
}
