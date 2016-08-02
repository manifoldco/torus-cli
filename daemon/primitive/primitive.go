package primitive

import (
	"time"

	"github.com/arigatomachine/cli/daemon/base64"
	"github.com/arigatomachine/cli/daemon/identity"
)

type User struct {
	Master *struct {
		Alg   string        `json:"alg"`
		Value *base64.Value `json:"value"`
	} `json:"master"`
}

func (u *User) Version() int {
	return 1
}

func (u *User) Type() byte {
	return byte(0x01)
}

// Signature, while not technically a payload, is still immutable, and must be
// orderer properly so that ID generation is correct.
//
// If PublicKeyID is nil, the signature is self-signed.
type Signature struct {
	Algorithm   string        `json:"alg"`
	PublicKeyID *identity.ID  `json:"public_key_id"`
	Value       *base64.Value `json:"value"`
}

// Immutable object payloads. Their fields must be lexicographically ordered by
// the json value, so we can correctly calculate the signature.

type PrivateKeyValue struct {
	Algorithm string        `json:"alg"`
	Value     *base64.Value `json:"value"`
}

type PrivateKey struct {
	Key         PrivateKeyValue `json:"key"`
	OrgID       *identity.ID    `json:"org_id"`
	OwnerID     *identity.ID    `json:"owner_id"`
	PNonce      *base64.Value   `json:"pnonce"`
	PublicKeyID *identity.ID    `json:"public_key_id"`
}

func (pk *PrivateKey) Version() int {
	return 1
}

func (pk *PrivateKey) Type() byte {
	return byte(0x07)
}

type PublicKeyValue struct {
	Value *base64.Value `json:"value"`
}

type PublicKey struct {
	Algorithm string         `json:"alg"`
	Created   time.Time      `json:"created_at"`
	Expires   time.Time      `json:"expires_at"`
	Key       PublicKeyValue `json:"key"`
	OrgID     *identity.ID   `json:"org_id"`
	OwnerID   *identity.ID   `json:"owner_id"`
	KeyType   string         `json:"type"`
}

func (pk *PublicKey) Version() int {
	return 1
}

func (pk *PublicKey) Type() byte {
	return byte(0x06)
}

const (
	SignatureClaimType  = "signature"
	RevocationClaimType = "revocation"
)

type Claim struct {
	Created     time.Time    `json:"created_at"`
	OrgID       *identity.ID `json:"org_id"`
	OwnerID     *identity.ID `json:"owner_id"`
	Previous    *identity.ID `json:"previous"`
	PublicKeyID *identity.ID `json:"public_key_id"`
	KeyType     string       `json:"type"`
}

func (c *Claim) Version() int {
	return 1
}

func (c *Claim) Type() byte {
	return byte(0x08)
}

// NewClaim returns a new Claim, with the created time set to now
func NewClaim(orgID, ownerID, previous, pubKeyID *identity.ID,
	keyType string) *Claim {
	return &Claim{
		OrgID:       orgID,
		OwnerID:     ownerID,
		Previous:    previous,
		PublicKeyID: pubKeyID,
		KeyType:     keyType,
		Created:     time.Now().UTC(),
	}
}
