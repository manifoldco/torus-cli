package apitypes

import (
	"errors"
	"fmt"

	"github.com/dchest/blake2b"

	"github.com/manifoldco/torus-cli/base32"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
)

// WorklogType is the enumerated byte type of WorklogItems
type WorklogType byte

// The enumberated byte types of WorklogItems
const (
	SecretRotateWorklogType WorklogType = 1 << iota
	MissingKeypairsWorklogType
	InviteApproveWorklogType
	UserKeyringMembersWorklogType
	MachineKeyringMembersWorklogType

	AnyWorklogType WorklogType = 0xff
)

// ErrIncorrectWorklogIDLen is returned when a base32 encoded worklog id is the
// wrong length.
var ErrIncorrectWorklogIDLen = errors.New("Incorrect worklog ID length")

const worklogIDLen = 9

// WorklogID is the unique content-based identifier for worklog entries
type WorklogID [worklogIDLen]byte

// DecodeWorklogIDFromString decodes a WorklogID from the given base32 encoded
// representation.
func DecodeWorklogIDFromString(raw string) (WorklogID, error) {
	id := WorklogID{}

	buf, err := base32.DecodeString(raw)
	if err != nil {
		return id, err
	}

	if len(buf) != worklogIDLen {
		return id, ErrIncorrectWorklogIDLen
	}

	copy(id[:], buf)
	return id, nil

}

func (id WorklogID) String() string {
	return base32.EncodeToString(id[:])
}

// Type returns this id's type
func (id WorklogID) Type() WorklogType {
	return WorklogType(id[0])
}

// WorklogItem is an item that the daemon has identified as needing to be done
// to ensure system correctness, or security in the face of stale secrets.
type WorklogItem struct {
	ID *WorklogID `json:"id"`

	Details WorklogDetails `json:"details"`
}

// Subject returns the human readable subject of this WorklogItem.
func (w *WorklogItem) Subject() string {
	return w.Details.Subject()
}

// Summary returns the human readable summary of this WorklogItem.
func (w *WorklogItem) Summary() string {
	return w.Details.Summary()
}

// WorklogDetails is the common interface exposed by worklog item types.
type WorklogDetails interface {
	Subject() string
	Summary() string
}

// InviteApproveWorklogDetails holds WorklogItem details for the
// InviteApproveWorklogType.
type InviteApproveWorklogDetails struct {
	InviteID *identity.ID `json:"invite_id"`
	Email    string       `json:"email"`
	Username string       `json:"username"`
	Name     string       `json:"name"`
	Org      string       `json:"org"`
	Teams    []string     `json:"teams"`
}

// Subject returns the human readable subject of this WorklogItem.
func (i *InviteApproveWorklogDetails) Subject() string {
	return i.Email
}

// Summary returns the human readable summary of this WorklogItem.
func (i *InviteApproveWorklogDetails) Summary() string {
	summary := fmt.Sprintf("The invite for %s to org %s is ready for approval.",
		i.Email, i.Org)
	return summary
}

// KeyringMembersWorklogDetails holds WorklogItem details for the
// KeyringMembersWorklogType.
type KeyringMembersWorklogDetails struct {
	EntityID *identity.ID      `json:"entity_id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	OwnerIDs []identity.ID     `json:"owner_ids"`
	Keyrings []pathexp.PathExp `json:"keyrings"`
}

// Subject returns the human readable subject of this WorklogItem.
func (k *KeyringMembersWorklogDetails) Subject() string {
	return k.Name
}

// Summary returns the human readable summary of this WorklogItem.
func (k *KeyringMembersWorklogDetails) Summary() string {
	return fmt.Sprintf("This %s is missing access to one or more secrets.", k.Type)
}

// MissingKeypairsWorklogDetails holds WorklogItem details for the
// MissingKeypairsWorklogType..
type MissingKeypairsWorklogDetails struct {
	Org               string `json:"org"`
	EncryptionMissing bool   `json:"encryption_missing"`
	SigningMissing    bool   `json:"signing_missing"`
}

// Subject returns the human readable subject of this WorklogItem.
func (m *MissingKeypairsWorklogDetails) Subject() string {
	return m.Org
}

// Summary returns the human readable summary of this WorklogItem.
func (m *MissingKeypairsWorklogDetails) Summary() string {
	msg := "Signing and encryption keypairs missing for org %s."
	if !m.EncryptionMissing {
		msg = "Signing keypair missing for org %s."
	} else if !m.SigningMissing {
		msg = "Encryption keypair missing for org %s."
	}

	return fmt.Sprintf(msg, m.Org)
}

// SecretRotateWorklogDetails holds WorklogItem details for the
// SecretRotateWorklogType.
type SecretRotateWorklogDetails struct {
	PathExp *pathexp.PathExp            `json:"pathexp"`
	Name    string                      `json:"name"`
	Reasons []SecretRotateWorklogReason `json:"reasons"`
}

// SecretRotateWorklogReason holds the username and claim revocation type
// for a secret rotation reason.
type SecretRotateWorklogReason struct {
	Username string                                `json:"username"`
	Type     primitive.KeyringMemberRevocationType `json:"type"`
}

// Subject returns the human readable subject of this WorklogItem.
func (s *SecretRotateWorklogDetails) Subject() string {
	return s.PathExp.String() + "/" + s.Name
}

// Summary returns the human readable summary of this WorklogItem.
func (s *SecretRotateWorklogDetails) Summary() string {
	return "A user's access was revoked. This secret's value should be changed."
}

// Type returns this item's type
func (w *WorklogItem) Type() WorklogType {
	return w.ID.Type()
}

// String returns a human reable string for this worklog item type.
func (t WorklogType) String() string {
	switch t {
	case SecretRotateWorklogType:
		return "secret"
	case MissingKeypairsWorklogType:
		return "keypairs"
	case InviteApproveWorklogType:
		return "invite"
	case UserKeyringMembersWorklogType:
		fallthrough
	case MachineKeyringMembersWorklogType:
		return "secret"
	default:
		return "n/a"
	}
}

// CreateID creates and populates a WorklogID for the WorklogItem based on the
// given type and its subject.
func (w *WorklogItem) CreateID(worklogType WorklogType) {
	h, err := blake2b.New(&blake2b.Config{Size: worklogIDLen - 1})
	if err != nil { // this only happens with a bad config
		panic(err)
	}

	h.Write([]byte{byte(worklogType)})
	h.Write([]byte(w.Details.Subject()))

	id := WorklogID{byte(worklogType)}
	copy(id[1:], h.Sum(nil))
	w.ID = &id
}
