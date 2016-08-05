// Package primitive contains definitions of the primitive types used
// in ag.
package primitive

import (
	"time"

	"github.com/arigatomachine/cli/daemon/base64"
	"github.com/arigatomachine/cli/daemon/identity"
)

// v1Schema embeds in other structs to indicate their schema version is 1.
type v1Schema struct{}

// Version returns the schema version of structs that embed this type.
func (v *v1Schema) Version() int {
	return 1
}

// User holds the details of a user, including their encrypted master key.
type User struct {
	v1Schema
	Master *struct {
		Alg   string        `json:"alg"`
		Value *base64.Value `json:"value"`
	} `json:"master"`
}

// Type returns the enumerated byte representation of User.
func (u *User) Type() byte {
	return byte(0x01)
}

// Signature is an immutable object, but not technically a payload. Its fields
// must be ordered properly so that ID generation is correct.
//
// If PublicKeyID is nil, the signature is self-signed.
type Signature struct {
	Algorithm   string        `json:"alg"`
	PublicKeyID *identity.ID  `json:"public_key_id"`
	Value       *base64.Value `json:"value"`
}

// Immutable object payloads. Their fields must be lexicographically ordered by
// the json value, so we can correctly calculate the signature.

// PrivateKeyValue holds the encrypted value of the PrivateKey.
type PrivateKeyValue struct {
	Algorithm string        `json:"alg"`
	Value     *base64.Value `json:"value"`
}

// PrivateKey is the private portion of an asymetric key.
type PrivateKey struct {
	v1Schema
	Key         PrivateKeyValue `json:"key"`
	OrgID       *identity.ID    `json:"org_id"`
	OwnerID     *identity.ID    `json:"owner_id"`
	PNonce      *base64.Value   `json:"pnonce"`
	PublicKeyID *identity.ID    `json:"public_key_id"`
}

// Type returns the enumerated byte representation of PrivateKey.
func (pk *PrivateKey) Type() byte {
	return byte(0x07)
}

// PublicKeyValue is the actual value of a PublicKey.
type PublicKeyValue struct {
	Value *base64.Value `json:"value"`
}

// PublicKey is the public portion of an asymetric key.
type PublicKey struct {
	v1Schema
	Algorithm string         `json:"alg"`
	Created   time.Time      `json:"created_at"`
	Expires   time.Time      `json:"expires_at"`
	Key       PublicKeyValue `json:"key"`
	OrgID     *identity.ID   `json:"org_id"`
	OwnerID   *identity.ID   `json:"owner_id"`
	KeyType   string         `json:"type"`
}

// Type returns the enumerated byte representation of PublicKey.
func (pk *PublicKey) Type() byte {
	return byte(0x06)
}

// Types of claims that can be made against public keys.
const (
	SignatureClaimType  = "signature"
	RevocationClaimType = "revocation"
)

// Claim is a signature or revocation claim against a public key.
type Claim struct {
	v1Schema
	Created     time.Time    `json:"created_at"`
	OrgID       *identity.ID `json:"org_id"`
	OwnerID     *identity.ID `json:"owner_id"`
	Previous    *identity.ID `json:"previous"`
	PublicKeyID *identity.ID `json:"public_key_id"`
	KeyType     string       `json:"type"`
}

// Type returns the enumerated byte representation of Claim.
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

// Credential is a secret value shared between a group of services based
// on users identity, operating environment, project, and organization
type Credential struct {
	v1Schema
	Name              string       `json:"name"`
	OrgID             *identity.ID `json:"org_id"`
	ProjectID         *identity.ID `json:"project_id"`
	PathExp           string       `json:"pathexp"`
	Previous          *identity.ID `json:"previous"`
	Value             string       `json:"value"`
	CredentialVersion int          `json:"version"`
}

// Type returns the enumerated byte representation of Credential
func (c *Credential) Type() byte {
	return byte(0xb)
}

// Org is a grouping of users that collaborate with each other
type Org struct {
	v1Schema
	Name string `json:"name"`
}

// Type returns the enumerated byte representation of Org
func (o *Org) Type() byte {
	return byte(0xd)
}
