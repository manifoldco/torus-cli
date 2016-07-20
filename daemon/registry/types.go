package registry

import (
	"encoding/base64"
	"errors"
	"reflect"
	"time"
)

type Error struct {
	StatusCode int

	Type string   `json:"type"`
	Err  []string `json:"error"`
}

func (e *Error) Error() string {
	return e.Type
}

type LoginTokenRequest struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}

type LoginTokenResponse struct {
	Salt  string `json:"salt"`
	Token string `json:"login_token"`
}

type AuthTokenRequest struct {
	Type      string `json:"type"`
	TokenHMAC string `json:"login_token_hmac"`
}

type AuthTokenResponse struct {
	Token string `json:"auth_token"`
}

type SelfResponse struct {
	ID      string `json:"ID"`
	Version uint8  `json:"version"`
	Body    *struct {
		Master *struct {
			Alg   string `json:"alg"`
			Value string `json:"value"`
		} `json:"master"`
	} `json:"body"`
}

type Base64Value []byte

func (bv *Base64Value) MarshalJSON() ([]byte, error) {
	return []byte("\"" + base64.RawURLEncoding.EncodeToString(*bv) + "\""), nil
}

func (bv *Base64Value) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != byte('"') || b[len(b)-1] != byte('"') {
		return errors.New("value is not a string")
	}

	out := make([]byte, base64.RawURLEncoding.DecodedLen(len(b)-2))
	n, err := base64.RawURLEncoding.Decode(out, b[1:len(b)-1])
	if err != nil {
		return err
	}

	v := reflect.ValueOf(bv).Elem()
	v.SetBytes(out[:n])
	return nil
}

// Envelope is the generic format for encapsulating request/response objects
// to/from arigato.
type Envelope struct {
	ID        *ID         `json:"id"`
	Version   uint8       `json:"version"`
	Body      interface{} `json:"body"`
	Signature Signature   `json:"sig"`
}

const (
	EncryptionKeyType = "encryption"
	SigningKeyType    = "signing"
)

// Immutable object payloads. Their fields must be lexicographically ordered by
// the json value, so we can correctly calculate the signature.

type PrivateKeyValue struct {
	Algorithm string       `json:"alg"`
	Value     *Base64Value `json:"value"`
}

type PrivateKey struct {
	Key         PrivateKeyValue `json:"key"`
	OrgID       *ID             `json:"org_id"`
	OwnerID     *ID             `json:"owner_id"`
	PNonce      *Base64Value    `json:"pnonce"`
	PublicKeyID *ID             `json:"public_key_id"`
}

func (pk *PrivateKey) Type() byte {
	return byte(0x07)
}

type PublicKeyValue struct {
	Value *Base64Value `json:"value"`
}
type PublicKey struct {
	Algorithm string         `json:"alg"`
	Created   time.Time      `json:"created_at"`
	Expires   time.Time      `json:"expires_at"`
	Key       PublicKeyValue `json:"key"`
	OrgID     *ID            `json:"org_id"`
	OwnerID   *ID            `json:"owner_id"`
	KeyType   string         `json:"type"`
}

func (pk *PublicKey) Type() byte {
	return byte(0x06)
}

// Signature, while not technically a payload, is still immutable, and must be
// orderer properly so that ID generation is correct.
//
// If PublicKeyID is nil, the signature is self-signed.
type Signature struct {
	Algorithm   string       `json:"alg"`
	PublicKeyID *ID          `json:"public_key_id"`
	Value       *Base64Value `json:"value"`
}
