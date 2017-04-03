// Package primitive contains definitions of the primitive types used
// in ag.
package primitive

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
)

// v1Schema embeds in other structs to indicate their schema version is 1.
type v1Schema struct{}

// Version returns the schema version of structs that embed this type.
func (v1Schema) Version() int {
	return 1
}

// Embedding the immutable or mutable struct in a struct denotes if it is
// immutable and should be signed, or a mutable type.

type immutable struct{}

func (immutable) Immutable() {}

type mutable struct{}

func (mutable) Mutable() {}

// v2Schema embeds in other structs to indicate their schema version is 2.
type v2Schema struct{}

// Version returns the schema version of structs that embed this type.
func (v2Schema) Version() int {
	return 2
}

// LoginPublicKey represents the public component of a asymmetric key used to
// authenticate against the registry
type LoginPublicKey struct {
	Alg   string        `json:"alg"`
	Salt  *base64.Value `json:"salt"`
	Value *base64.Value `json:"value"`
}

// BaseUser represents the common properties shared between all user schema
// versions.
type BaseUser struct {
	mutable
	Username string        `json:"username"`
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	State    string        `json:"state"`
	Password *UserPassword `json:"password"`
	Master   *MasterKey    `json:"master"`
}

// User is the body of a user object
type User struct { // type: 0x01
	v2Schema
	BaseUser
	PublicKey *LoginPublicKey `json:"public_key"`
}

// UserV1 is the body of a user object
type UserV1 struct { // type: 0x01
	v1Schema
	BaseUser
}

// MasterKey is the body.master object for a user and machine token
type MasterKey struct {
	Value *base64.Value `json:"value"`
	Alg   string        `json:"alg"`
}

// UserPassword is the body.password object for a user
type UserPassword struct {
	Salt  string        `json:"salt"`
	Value *base64.Value `json:"value"`
	Alg   string        `json:"alg"`
}

// TokenType represents the different types of tokens
type TokenType string

// Types of tokens which are created throughout the authentication flow
const (
	LoginToken TokenType = "login"
	AuthToken  TokenType = "auth"
)

// AuthMechanism represents the different authentication mechanisms used for
// granting Tokens of type AuthToken
type AuthMechanism string

// Types of mechanisms used to authenticate a user or machine
const (
	HMACAuth         AuthMechanism = "hmac"
	EdDSAAuth        AuthMechanism = "eddsa"
	UpgradeEdDSAAuth AuthMechanism = "upgrade-eddsa"
)

// Token is the body of a token object
type Token struct { // type: 0x10
	v2Schema
	mutable
	TokenType TokenType     `json:"type"`
	Token     string        `json:"token"`
	OwnerID   *identity.ID  `json:"owner_id"`
	Mechanism AuthMechanism `json:"mechanism"`
}

// Signature is an immutable object, but not technically a payload.
// If PublicKeyID is nil, the signature is self-signed.
type Signature struct {
	Algorithm   string        `json:"alg"`
	PublicKeyID *identity.ID  `json:"public_key_id"`
	Value       *base64.Value `json:"value"`
}

// PrivateKeyValue holds the encrypted value of the PrivateKey.
type PrivateKeyValue struct {
	Algorithm string        `json:"alg"`
	Value     *base64.Value `json:"value"`
}

// PrivateKey is the private portion of an asymetric key.
type PrivateKey struct { // type: 0x07
	v1Schema
	immutable
	Key         PrivateKeyValue `json:"key"`
	OrgID       *identity.ID    `json:"org_id"`
	OwnerID     *identity.ID    `json:"owner_id"`
	PNonce      *base64.Value   `json:"pnonce"`
	PublicKeyID *identity.ID    `json:"public_key_id"`
}

// PublicKeyValue is the actual value of a PublicKey.
type PublicKeyValue struct {
	Value *base64.Value `json:"value"`
}

// KeyType the enumeration of all types of keys.
type KeyType string

// Types of keys supported by the system.
const (
	EncryptionKeyType KeyType = "encryption"
	SigningKeyType    KeyType = "signing"
)

// PublicKey is the public portion of an asymetric key.
type PublicKey struct { // type: 0x06
	v1Schema
	immutable
	Algorithm string         `json:"alg"`
	Created   time.Time      `json:"created_at"`
	Expires   time.Time      `json:"expires_at"`
	Key       PublicKeyValue `json:"key"`
	OrgID     *identity.ID   `json:"org_id"`
	OwnerID   *identity.ID   `json:"owner_id"`
	KeyType   KeyType        `json:"type"`
}

// ClaimType is the enumeration of all claims that can be made against public
// keys.
type ClaimType string

// Types of claims that can be made against public keys.
const (
	SignatureClaimType  ClaimType = "signature"
	RevocationClaimType ClaimType = "revocation"
)

// Claim is a signature or revocation claim against a public key.
type Claim struct { // type: 0x08
	v1Schema
	immutable
	Created     time.Time    `json:"created_at"`
	OrgID       *identity.ID `json:"org_id"`
	OwnerID     *identity.ID `json:"owner_id"`
	Previous    *identity.ID `json:"previous"`
	PublicKeyID *identity.ID `json:"public_key_id"`
	ClaimType   ClaimType    `json:"type"`
}

// NewClaim returns a new Claim, with the created time set to now
func NewClaim(orgID, ownerID, previous, pubKeyID *identity.ID, claimType ClaimType) *Claim {
	return &Claim{
		OrgID:       orgID,
		OwnerID:     ownerID,
		Previous:    previous,
		PublicKeyID: pubKeyID,
		ClaimType:   claimType,
		Created:     time.Now().UTC(),
	}
}

// Credential is a secret value shared between a group of services based
// on users identity, operating environment, project, and organization
type Credential struct { // type: 0x0b
	v2Schema
	immutable
	BaseCredential
	State *string `json:"state"`
}

// CredentialV1 is a secret value shared between a group of services based
// on users identity, operating environment, project, and organization
type CredentialV1 struct { // type: 0x0b
	v1Schema
	immutable
	BaseCredential
}

// BaseCredential is a secret value shared between a group of services based
// on users identity, operating environment, project, and organization
type BaseCredential struct {
	Credential        *CredentialValue `json:"credential"`
	KeyringID         *identity.ID     `json:"keyring_id"`
	Name              string           `json:"name"`
	Nonce             *base64.Value    `json:"nonce"`
	OrgID             *identity.ID     `json:"org_id"`
	PathExp           *pathexp.PathExp `json:"pathexp"`
	Previous          *identity.ID     `json:"previous"`
	ProjectID         *identity.ID     `json:"project_id"`
	CredentialVersion int              `json:"version"`
}

// CredentialValue is the secretbox encrypted value of the containing
// Credential.
type CredentialValue struct {
	Algorithm string        `json:"alg"`
	Nonce     *base64.Value `json:"nonce"`
	Value     *base64.Value `json:"value"`
}

// BaseKeyring is the shared structure between keyring schema versions.
type BaseKeyring struct {
	immutable
	Created        time.Time        `json:"created_at"`
	OrgID          *identity.ID     `json:"org_id"`
	PathExp        *pathexp.PathExp `json:"pathexp"`
	Previous       *identity.ID     `json:"previous"`
	ProjectID      *identity.ID     `json:"project_id"`
	KeyringVersion int              `json:"version"`
}

// KeyringV1 is the old keyring format, without claims or mekshares.
type KeyringV1 struct { // type: 0x09
	v1Schema
	BaseKeyring
}

// Keyring is a mechanism for sharing a shared secret between many different
// users and machines at a position in the credential path.
//
// Credentials belong to Keyrings
type Keyring struct { // type: 0x09
	v2Schema
	BaseKeyring
}

// NewKeyring returns a new v2 Keyring, with the created time set to now
func NewKeyring(orgID, projectID *identity.ID, pathExp *pathexp.PathExp) *Keyring {
	return &Keyring{
		BaseKeyring: BaseKeyring{
			Created:   time.Now().UTC(),
			OrgID:     orgID,
			PathExp:   pathExp,
			ProjectID: projectID,

			// This is the first instance of the keyring, so version is 1,
			// and there is no previous instance.
			Previous:       nil,
			KeyringVersion: 1,
		},
	}
}

// KeyringMemberV1 is a record of sharing a master secret key with a user or
// machine.
//
// KeyringMember belongs to a Keyring
type KeyringMemberV1 struct { // type: 0x0a
	v1Schema
	immutable
	Created         time.Time         `json:"created_at"`
	EncryptingKeyID *identity.ID      `json:"encrypting_key_id"`
	Key             *KeyringMemberKey `json:"key"`
	KeyringID       *identity.ID      `json:"keyring_id"`
	OrgID           *identity.ID      `json:"org_id"`
	OwnerID         *identity.ID      `json:"owner_id"`
	ProjectID       *identity.ID      `json:"project_id"`
	PublicKeyID     *identity.ID      `json:"public_key_id"`
}

// KeyringMember is a record of sharing a master secret key with a user or
// machine.
//
// This is the v2 schema version, which has a detached mekshare so it can be
// revoked.
//
// KeyringMember belongs to a Keyring
type KeyringMember struct { // type: 0x0a
	v2Schema
	immutable
	Created         time.Time    `json:"created_at"`
	EncryptingKeyID *identity.ID `json:"encrypting_key_id"`
	KeyringID       *identity.ID `json:"keyring_id"`
	OrgID           *identity.ID `json:"org_id"`
	OwnerID         *identity.ID `json:"owner_id"`
	PublicKeyID     *identity.ID `json:"public_key_id"`
}

// KeyringMemberKey is the keyring master encryption key, encrypted for the
// owner of a KeyringMember/MEKShare
type KeyringMemberKey struct {
	Algorithm string        `json:"alg"`
	Nonce     *base64.Value `json:"nonce"`
	Value     *base64.Value `json:"value"`
}

// MEKShare is a V2 KeyringMember's share of the keyring master encryption key.
type MEKShare struct { // type: 0x16
	v1Schema
	immutable
	Created         time.Time         `json:"created_at"`
	OrgID           *identity.ID      `json:"org_id"`
	OwnerID         *identity.ID      `json:"owner_id"`
	KeyringID       *identity.ID      `json:"keyring_id"`
	KeyringMemberID *identity.ID      `json:"keyring_member_id"`
	Key             *KeyringMemberKey `json:"key"`
}

// KeyringMemberClaim is a claim for a keyring member. Only revocation is
// supported as a claim type.
type KeyringMemberClaim struct { // type: 0x15
	v1Schema
	immutable
	OrgID           *identity.ID              `json:"org_id"`
	KeyringID       *identity.ID              `json:"keyring_id"`
	KeyringMemberID *identity.ID              `json:"keyring_member_id"`
	OwnerID         *identity.ID              `json:"owner_id"`
	Previous        *identity.ID              `json:"previous"`
	ClaimType       ClaimType                 `json:"type"`
	Reason          *KeyringMemberClaimReason `json:"reason"`
	Created         time.Time                 `json:"created_at"`
}

// KeyringMemberClaimReason holds the type and optional details of the reason
// for a KeyringMember's revocation.
type KeyringMemberClaimReason struct {
	Type   KeyringMemberRevocationType   `json:"type"`
	Params KeyringMemberRevocationParams `json:"params"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (k *KeyringMemberClaimReason) UnmarshalJSON(b []byte) error {

	raw := struct {
		Type   KeyringMemberRevocationType `json:"type"`
		Params json.RawMessage             `json:"params"`
	}{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	k.Type = raw.Type

	switch k.Type {
	case OrgRemovalRevocationType: // no params
		return nil
	case KeyRevocationRevocationType:
		k.Params = &KeyRevocationRevocationParams{}
	case MachineDestroyRevocationType:
		k.Params = &MachineDestroyRevocationParams{}
	case MachineTokenDestroyRevocationType:
		k.Params = &MachineTokenDestroyRevocationParams{}
	}

	return json.Unmarshal(raw.Params, k.Params)
}

// KeyringMemberRevocationType is the enumerated byte type of keyring membership
// revocation reasons.
type KeyringMemberRevocationType byte

// The keyring membership revocation reasons.
const (
	OrgRemovalRevocationType KeyringMemberRevocationType = iota
	KeyRevocationRevocationType
	MachineDestroyRevocationType
	MachineTokenDestroyRevocationType
)

func (k KeyringMemberRevocationType) String() string {
	switch k {
	case OrgRemovalRevocationType:
		return "org_removal"
	case KeyRevocationRevocationType:
		return "key_revocation"
	case MachineDestroyRevocationType:
		return "machine_destroy"
	case MachineTokenDestroyRevocationType:
		return "machine_token_destroy"
	default:
		panic("invalid revocation type value")
	}
}

// MarshalText implements the encoding.TextMarshaler interface, used for JSON
// marshaling.
func (k KeyringMemberRevocationType) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

var errBadRevocationType = errors.New("unknown revocation type")

// UnmarshalText implements the encoding.TextUnmarshaler interface, used for
// JSON unmarshaling.
func (k *KeyringMemberRevocationType) UnmarshalText(b []byte) error {
	switch string(b) {
	case "org_removal":
		*k = OrgRemovalRevocationType
	case "key_revocation":
		*k = KeyRevocationRevocationType
	case "machine_destroy":
		*k = MachineDestroyRevocationType
	case "machine_token_destroy":
		*k = MachineTokenDestroyRevocationType
	default:
		return errBadRevocationType
	}

	return nil
}

// KeyringMemberRevocationParams is the interface for holding additional details
// about a membership revocation, based on the reason type.
type KeyringMemberRevocationParams interface{}

// KeyRevocationRevocationParams holds details for a key_revocation revocation
// type.
type KeyRevocationRevocationParams struct {
	PublicKeyID *identity.ID `json:"public_key_id"`
}

// MachineDestroyRevocationParams holds details for a machine_destroy revocation
// type.
type MachineDestroyRevocationParams struct {
	MachineID *identity.ID `json:"machine_id"`
}

// MachineTokenDestroyRevocationParams holds details for a machine_token_destroy
// revocation type.
type MachineTokenDestroyRevocationParams struct {
	MachineTokenID *identity.ID `json:"machine_token_id"`
}

// Org is a grouping of users that collaborate with each other
type Org struct { // type: 0x0d
	v1Schema
	mutable
	Name string `json:"name"`
}

// Org Invitations exist in four states: pending, associated,
// accepted, and approved.
const (
	OrgInvitePendingState    = "pending"
	OrgInviteAssociatedState = "associated"
	OrgInviteAcceptedState   = "accepted"
	OrgInviteApprovedState   = "approved"
)

// OrgInvite is an invitation for an individual to join an organization
type OrgInvite struct { // type: 0x13
	v1Schema
	mutable
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

// Machines can be in one of two states: active or destroyed
const (
	MachineActiveState    = "active"
	MachineDestroyedState = "destroyed"
)

// Machine is an entity that represents a machine object
type Machine struct { // type: 0x17
	v1Schema
	mutable
	Name        string       `json:"name"`
	OrgID       *identity.ID `json:"org_id"`
	CreatedBy   *identity.ID `json:"created_by"`
	Created     time.Time    `json:"created_at"`
	DestroyedBy *identity.ID `json:"destroyed_by"`
	Destroyed   *time.Time   `json:"destroyed_at"`
	State       string       `json:"state"`
}

// MachineTokens can be in one of two states: active or destroyed
const (
	MachineTokenActiveState    = "active"
	MachineTokenDestroyedState = "destroyed"
)

// MachineToken is an portion of the MachineSegment object
type MachineToken struct { // type: 0x18
	v1Schema
	mutable
	OrgID       *identity.ID    `json:"org_id"`
	MachineID   *identity.ID    `json:"machine_id"`
	PublicKey   *LoginPublicKey `json:"public_key"`
	Master      *MasterKey      `json:"master"`
	CreatedBy   *identity.ID    `json:"created_by"`
	Created     time.Time       `json:"created_at"`
	DestroyedBy *identity.ID    `json:"destroyed_by"`
	Destroyed   *time.Time      `json:"destroyed_at"`
	State       string          `json:"state"`
}

// Project is an entity that represents a group of services
type Project struct { // type: 0x04
	v1Schema
	mutable
	Name  string       `json:"name"`
	OrgID *identity.ID `json:"org_id"`
}

// Policy is an entity that represents a group of statements for acl
type Policy struct { // type: 0x11
	v1Schema
	mutable
	PolicyType string       `json:"type"`
	Previous   *identity.ID `json:"previous"`
	OrgID      *identity.ID `json:"org_id"`
	Policy     struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Statements  []PolicyStatement `json:"statements"`
	} `json:"policy"`
}

// PolicyStatement is an acl statement on a policy object
type PolicyStatement struct {
	Effect   PolicyEffect `json:"effect"`
	Action   PolicyAction `json:"action"`
	Resource string       `json:"resource"`
}

// PolicyEffect is the effect type of the statement (allow or deny)
type PolicyEffect bool

// These are the two policy effect types
const (
	PolicyEffectAllow = true
	PolicyEffectDeny  = false
)

// MarshalText implements the encoding.TextMarshaler interface, used for JSON
// marshaling.
func (pe *PolicyEffect) MarshalText() ([]byte, error) {
	if *pe {
		return []byte("allow"), nil
	}
	return []byte("deny"), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface, used for
// JSON unmarshaling.
func (pe *PolicyEffect) UnmarshalText(b []byte) error {
	*pe = string(b) == "allow" || string(b) == "sudo"
	return nil
}

// String returns a string representation of the PolicyEffect (allow or deny)
func (pe *PolicyEffect) String() string {
	b, _ := pe.MarshalText()
	return string(b)
}

// PolicyAction represents the user actions that are covered by a statement.
type PolicyAction byte

// These are all the possible PolicyActions
const (
	PolicyActionCreate = 1 << iota
	PolicyActionRead
	PolicyActionUpdate
	PolicyActionDelete
	PolicyActionList
)

var policyActionStrings = []string{
	"create",
	"read",
	"update",
	"delete",
	"list",
}

// MarshalJSON implements the json.Marshaler interface. A PolicyAction is
// encoded in JSON either the string representations of its actions in a list,
// or a single string when there is only one action.
func (pa *PolicyAction) MarshalJSON() ([]byte, error) {
	out := []string{}

	for i, v := range policyActionStrings {
		if (1<<uint(i))&byte(*pa) > 0 {
			out = append(out, v)
		}
	}

	if len(out) == 1 {
		return json.Marshal(out[0])
	}

	return json.Marshal(out)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pa *PolicyAction) UnmarshalJSON(b []byte) error {
	raw := make([]string, 1)

	var err error
	if len(b) > 0 && b[0] == '"' {
		err = json.Unmarshal(b, &raw[0])
	} else {
		err = json.Unmarshal(b, &raw)
	}

	if err != nil {
		return err
	}

	for _, a := range raw {
		for i, v := range policyActionStrings {
			if a == v {
				*pa |= 1 << uint(i)
				continue
			}
		}
	}

	return nil
}

func (pa *PolicyAction) String() string {
	out := []string{}

	for i, v := range policyActionStrings {
		if (1<<uint(i))&byte(*pa) > 0 {
			out = append(out, v)
		}
	}

	return strings.Join(out, ", ")
}

// ShortString displays a single character representation of each of the
// policy's actions.
func (pa *PolicyAction) ShortString() string {
	out := []byte{}

	for i, v := range policyActionStrings {
		if (1<<uint(i))&byte(*pa) > 0 {
			out = append(out, v[0])
		} else {
			out = append(out, '-')
		}
	}

	return string(out)
}

// PolicyAttachment is an entity that represents the link between policies and teams
type PolicyAttachment struct { // type: 0x12
	v1Schema
	mutable
	OwnerID  *identity.ID `json:"owner_id"`
	PolicyID *identity.ID `json:"policy_id"`
	OrgID    *identity.ID `json:"org_id"`
}

// Service is an entity that represents a group of processes
type Service struct { // type: 0x03
	v1Schema
	mutable
	Name      string       `json:"name"`
	OrgID     *identity.ID `json:"org_id"`
	ProjectID *identity.ID `json:"project_id"`
}

// Environment is an entity that represents a group of processes
type Environment struct { // type: 0x05
	v1Schema
	mutable
	Name      string       `json:"name"`
	OrgID     *identity.ID `json:"org_id"`
	ProjectID *identity.ID `json:"project_id"`
}

// TeamType is the type that holds the enumeration of possible team types.
type TeamType string

// There are three types of teams: system, machine and user. System teams are
// managed by the Torus registry while Machine teams contain only machines.
const (
	AnyTeamType     TeamType = ""
	SystemTeamType  TeamType = "system"
	UserTeamType    TeamType = "user"
	MachineTeamType TeamType = "machine"
)

// Teams are used to represent a group of identities and their associated
// access control policies
const (
	AdminTeamName   = "admin"
	OwnerTeamName   = "owner"
	MemberTeamName  = "member"
	MachineTeamName = "machine"
)

// Team IDs for certain system teams can be derived based on their OrgID.
const (
	DerivableMachineTeamSymbol = 0x04
)

// Team is an entity that represents a group of users
type Team struct { // type: 0x0f
	v1Schema
	mutable
	Name     string       `json:"name"`
	OrgID    *identity.ID `json:"org_id"`
	TeamType TeamType     `json:"type"`
}

// Membership is an entity that represents whether a user or
// machine is a part of a team in an organization.
type Membership struct { // type: 0x0e
	v1Schema
	mutable
	OrgID   *identity.ID `json:"org_id"`
	OwnerID *identity.ID `json:"owner_id"`
	TeamID  *identity.ID `json:"team_id"`
}
