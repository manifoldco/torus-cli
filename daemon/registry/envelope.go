package registry

import (
	"encoding/json"
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
