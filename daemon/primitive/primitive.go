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
	Credential        *CredentialValue `json:"credential"`
	KeyringID         *identity.ID     `json:"keyring_id"`
	Name              string           `json:"name"`
	Nonce             *base64.Value    `json:"nonce"`
	OrgID             *identity.ID     `json:"org_id"`
	PathExp           string           `json:"pathexp"`
	Previous          *identity.ID     `json:"previous"`
	ProjectID         *identity.ID     `json:"project_id"`
	CredentialVersion int              `json:"version"`
}

// Type returns the enumerated byte representation of Credential
func (c *Credential) Type() byte {
	return byte(0xb)
}

// CredentialValue is the secretbox encrypted value of the containing
// Credential.
type CredentialValue struct {
	Algorithm string        `json:"alg"`
	Nonce     *base64.Value `json:"nonce"`
	Value     *base64.Value `json:"value"`
}

// Keyring is a mechanism for sharing a shared secret between many different
// users and machines at a position in the credential path.
//
// Credentials belong to Keyrings
type Keyring struct {
	v1Schema
	Created        time.Time    `json:"created_at"`
	OrgID          *identity.ID `json:"org_id"`
	PathExp        string       `json:"pathexp"`
	Previous       *identity.ID `json:"previous"`
	ProjectID      *identity.ID `json:"project_id"`
	KeyringVersion int          `json:"version"`
}

// Type returns the enumerated byte representation of Keyring
func (k *Keyring) Type() byte {
	return byte(0x09)
}

// KeyringMember is a record of sharing a master secret key with a user or
// machine.
//
// KeyringMember belongs to a Keyring
type KeyringMember struct {
	v1Schema
	Created         time.Time         `json:"created_at"`
	EncryptingKeyID *identity.ID      `json:"encrypting_key_id"`
	Key             *KeyringMemberKey `json:"key"`
	KeyringID       *identity.ID      `json:"keyring_id"`
	OrgID           *identity.ID      `json:"org_id"`
	OwnerID         *identity.ID      `json:"owner_id"`
	ProjectID       *identity.ID      `json:"project_id"`
	PublicKeyID     *identity.ID      `json:"public_key_id"`
}

// Type returns the enumerated byte representation of KeyringMember
func (km *KeyringMember) Type() byte {
	return byte(0x0a)
}

// KeyringMemberKey is the keyring master encryption key, encrypted for the
// owner of a KeyringMember
type KeyringMemberKey struct {
	Algorithm string        `json:"alg"`
	Nonce     *base64.Value `json:"nonce"`
	Value     *base64.Value `json:"value"`
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

// OrgInvite is an invitation for an individual to join an organization
type OrgInvite struct {
	v1Schema
	OrgID      *identity.ID `json:"org_id"`
	Email      string       `json:"email"`
	InviterID  *identity.ID `json:"inviter_id"`
	InviteeID  *identity.ID `json:"invitee_id"`
	ApproverID *identity.ID `json:"approver_id"`
	State      string       `json:"state"`
	Code       *struct {
		Alg   string        `json:"alg"`
		Salt  *base64.Value `json:"salt"`
		Value *base64.Value `json:"value"`
	} `json:"code"`
	PendingTeams []identity.ID `json:"pending_teams"`
	Created      *time.Time    `json:"created_at"`
	Accepted     *time.Time    `json:"accepted_at"`
	Approved     *time.Time    `json:"approved_at"`
}

// Type returns the numerated byte representation of OrgInvite
func (o *OrgInvite) Type() byte {
	return byte(0x13)
}
