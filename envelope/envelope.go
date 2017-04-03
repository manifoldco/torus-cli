// Package envelope defines the generic encapsulating format for torus
// objects.
package envelope

import (
	"fmt"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
)

// Envelope is the interface implemented by objects that encapsulate 'true'
// torus objects.
type Envelope interface {
	GetID() *identity.ID // avoid field collision
}

// Signed is the generic format for encapsulating signed immutable
// request/response objects to/from torus.
type Signed struct {
	ID        *identity.ID        `json:"id"`
	Version   uint8               `json:"version"`
	Body      identity.Immutable  `json:"body"`
	Signature primitive.Signature `json:"sig"`
}

// GetID returns the ID of the object encapsulated in this envelope.
func (e *Signed) GetID() *identity.ID {
	return e.ID
}

// Unsigned is the generic format for encapsulating unsigned mutable
// request/response objects to/from torus.
type Unsigned struct {
	ID      *identity.ID     `json:"id"`
	Version uint8            `json:"version"`
	Body    identity.Mutable `json:"body"`
}

// GetID returns the ID of the object encapsulated in this envelope.
func (e *Unsigned) GetID() *identity.ID {
	return e.ID
}

// GetVersion returns the Schema version of the object encapsulated in this
// envelope.
func (e *Unsigned) GetVersion() uint8 {
	return e.Version
}

// UserInf is the common interface for all user schema versions.
type UserInf interface {
	Envelope
	Username() string
	Name() string
	Email() string
	Password() *primitive.UserPassword
	Master() *primitive.MasterKey
}

// Username returns the username owned by this User.
func (u *UserV1) Username() string {
	return u.Body.Username
}

// Name returns the Full Name of this User.
func (u *UserV1) Name() string {
	return u.Body.Name
}

// Email returns the current Email Address for this User.
func (u *UserV1) Email() string {
	return u.Body.Email
}

// Password returns the UserPassword for this User.
func (u *UserV1) Password() *primitive.UserPassword {
	return u.Body.Password
}

// Master returns the MasterKey for this User.
func (u *UserV1) Master() *primitive.MasterKey {
	return u.Body.Master
}

// Username returns the username owned by this User.
func (u *User) Username() string {
	return u.Body.Username
}

// Name returns the Full Name of this User.
func (u *User) Name() string {
	return u.Body.Name
}

// Email returns the current Email Address for this User.
func (u *User) Email() string {
	return u.Body.Email
}

// Password returns the UserPassword for this User.
func (u *User) Password() *primitive.UserPassword {
	return u.Body.Password
}

// Master returns the MasterKey for this User.
func (u *User) Master() *primitive.MasterKey {
	return u.Body.Master
}

// ConvertUser converts an unsigned envelope to a UserInf interface which
// provides a common interface for all user versions.
func ConvertUser(e *Unsigned) (UserInf, error) {
	var user UserInf
	switch e.Version {
	case 1:
		user = &UserV1{
			ID:      e.ID,
			Version: e.Version,
			Body:    e.Body.(*primitive.UserV1),
		}
	case 2:
		user = &User{
			ID:      e.ID,
			Version: e.Version,
			Body:    e.Body.(*primitive.User),
		}
	default:
		return nil, fmt.Errorf("Unknown User Schema Version: %d", e.Version)
	}

	return user, nil
}

// KeyringInf is the common interface for all keyring schema versions.
type KeyringInf interface {
	Envelope
	PathExp() *pathexp.PathExp
	OrgID() *identity.ID
	GetVersion() uint8 // Return the schema version of the keyring
}

// GetVersion returns the schema version of this Keyring.
func (k *KeyringV1) GetVersion() uint8 {
	return k.Version
}

// OrgID returns the ID of the Org that this Keyring belongs to.
func (k *KeyringV1) OrgID() *identity.ID {
	return k.Body.OrgID
}

// PathExp returns the path expression that contains all Credentials in this
// Keyring.
func (k *KeyringV1) PathExp() *pathexp.PathExp {
	return k.Body.PathExp
}

// GetVersion returns the schema version of this Keyring.
func (k *Keyring) GetVersion() uint8 {
	return k.Version
}

// OrgID returns the ID of the Org that this Keyring belongs to.
func (k *Keyring) OrgID() *identity.ID {
	return k.Body.OrgID
}

// PathExp returns the path expression that contains all Credentials in this
// Keyring.
func (k *Keyring) PathExp() *pathexp.PathExp {
	return k.Body.PathExp
}

// CredentialInf is the common interface for all Credential schema versions.
type CredentialInf interface {
	Envelope
	GetVersion() uint8 // schema version

	Previous() *identity.ID
	CredentialVersion() int

	PathExp() *pathexp.PathExp
	Name() string

	Unset() bool
	Nonce() *base64.Value
	Credential() *primitive.CredentialValue

	OrgID() *identity.ID
	ProjectID() *identity.ID
}

// GetVersion returns the schema version of this Credential.
func (c *CredentialV1) GetVersion() uint8 {
	return c.Version
}

// Previous returns the ID of the previous versino of this Credential, or nil
// if this Credential has no previous version.
func (c *CredentialV1) Previous() *identity.ID {
	return c.Body.Previous
}

// CredentialVersion returns the monotomically incremented version of the
// Credential for this PathExp/Name pair.
func (c *CredentialV1) CredentialVersion() int {
	return c.Body.CredentialVersion
}

// PathExp returns the path expression for this Credential's location.
func (c *CredentialV1) PathExp() *pathexp.PathExp {
	return c.Body.PathExp
}

// Name returns this Credential's name.
func (c *CredentialV1) Name() string {
	return c.Body.Name
}

// Unset returns a bool indicating if this Credential has been explicitly unset.
// Version 1 credentials do not track this, so it is always false.
func (CredentialV1) Unset() bool {
	return false
}

// Nonce returns the Nonce for this Credential's encrypted value.
func (c *CredentialV1) Nonce() *base64.Value {
	return c.Body.Nonce
}

// Credential returns the encrypted CredentialValue for this Credential.
func (c *CredentialV1) Credential() *primitive.CredentialValue {
	return c.Body.Credential
}

// OrgID returns the ID of the Org that this Credential belongs to.
func (c *CredentialV1) OrgID() *identity.ID {
	return c.Body.OrgID
}

// ProjectID returns the ID of the Project that this Credential belongs to.
func (c *CredentialV1) ProjectID() *identity.ID {
	return c.Body.ProjectID
}

// GetVersion returns the schema version of this Credential.
func (c *Credential) GetVersion() uint8 {
	return c.Version
}

// Previous returns the ID of the previous versino of this Credential, or nil
// if this Credential has no previous version.
func (c *Credential) Previous() *identity.ID {
	return c.Body.Previous
}

// CredentialVersion returns the monotomically incremented version of the
// Credential for this PathExp/Name pair.
func (c *Credential) CredentialVersion() int {
	return c.Body.CredentialVersion
}

// PathExp returns the path expression for this Credential's location.
func (c *Credential) PathExp() *pathexp.PathExp {
	return c.Body.PathExp
}

// Name returns this Credential's name.
func (c *Credential) Name() string {
	return c.Body.Name
}

// Unset returns a bool indicating if this Credential has been explicitly unset.
func (c *Credential) Unset() bool {
	return c.Body.State != nil && *c.Body.State == "unset"
}

// Nonce returns the Nonce for this Credential's encrypted value.
func (c *Credential) Nonce() *base64.Value {
	return c.Body.Nonce
}

// Credential returns the encrypted CredentialValue for this Credential.
func (c *Credential) Credential() *primitive.CredentialValue {
	return c.Body.Credential
}

// OrgID returns the ID of the Org that this Credential belongs to.
func (c *Credential) OrgID() *identity.ID {
	return c.Body.OrgID
}

// ProjectID returns the ID of the Project that this Credential belongs to.
func (c *Credential) ProjectID() *identity.ID {
	return c.Body.ProjectID
}
