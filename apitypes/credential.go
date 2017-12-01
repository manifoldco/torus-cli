package apitypes

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
)

var errMistmatchedType = errors.New("Mismatched type and value in credential")

const (
	unsetCV = iota
	stringCV
	intCV
	floatCV
	undecryptedCV // only used internally to the daemon
)

// CredentialEnvelope is an unencrypted credential object with a
// deserialized body
type CredentialEnvelope struct {
	ID      *identity.ID `json:"id"`
	Version uint8        `json:"version"`
	Body    *Credential  `json:"body"`
}

// CredentialResp is used to facilitate unmarshalling of versioned objects
type CredentialResp struct {
	ID      *identity.ID    `json:"id"`
	Version uint8           `json:"version"`
	Body    json.RawMessage `json:"body"`
}

// Credential interface is either a v1 or v2 credential object
type Credential interface {
	GetName() string
	GetOrgID() *identity.ID
	GetPathExp() *pathexp.PathExp
	GetProjectID() *identity.ID
	GetValue() *CredentialValue
}

// BaseCredential is the body of an unencrypted Credential
type BaseCredential struct {
	Name      string           `json:"name"`
	OrgID     *identity.ID     `json:"org_id"`
	PathExp   *pathexp.PathExp `json:"pathexp"`
	ProjectID *identity.ID     `json:"project_id"`
	Value     *CredentialValue `json:"value"`
}

// GetName returns the name
func (c *BaseCredential) GetName() string {
	return c.Name
}

// GetOrgID returns the org id
func (c *BaseCredential) GetOrgID() *identity.ID {
	return c.OrgID
}

// GetPathExp returns the pathexp
func (c *BaseCredential) GetPathExp() *pathexp.PathExp {
	return c.PathExp
}

// GetProjectID returns the project id
func (c *BaseCredential) GetProjectID() *identity.ID {
	return c.ProjectID
}

// GetValue returns the value object, unless unset then returns nil
func (c *BaseCredential) GetValue() *CredentialValue {
	if c.Value.cvtype == unsetCV {
		return nil
	}

	return c.Value
}

// CredentialV2 is the body of an unencrypted Credential
type CredentialV2 struct {
	BaseCredential
	State string `json:"state"`
}

// GetValue returns the value object, unless unset then returns nil
func (c *CredentialV2) GetValue() *CredentialValue {
	if c.State == "unset" { // v2 unset state
		return nil
	}
	if c.Value == nil { // v2 value nilled
		return nil
	}
	if c.Value.IsUnset() { // value contains v1 undefined type
		return nil
	}
	return c.Value
}

// CredentialValue is the raw value of a credential.
type CredentialValue struct {
	cvtype int
	value  string
	raw    interface{}
}

// IsUnset returns if this credential has been unset (deleted)
func (c *CredentialValue) IsUnset() bool {
	return c.cvtype == unsetCV
}

// IsUndecrypted returns if this credential has not been decrypted
func (c *CredentialValue) IsUndecrypted() bool {
	return c.cvtype == undecryptedCV
}

// String returns the string representation of this credential. It panics
// if the credential was deleted.
func (c *CredentialValue) String() string {
	if c.cvtype == unsetCV {
		panic("CredentialValue has been unset")
	}

	if c.cvtype == undecryptedCV {
		panic("CredentialValue was not decrypted")
	}

	return c.value
}

type credentialImpl struct {
	Version uint8 `json:"version"`
	Body    struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	} `json:"body"`
}

// Raw returns the underlying typed value for this Credential.
func (c *CredentialValue) Raw() (interface{}, error) {
	if c.IsUnset() {
		return nil, errors.New("Cannot return raw value of an unset Credential")
	}

	return c.raw, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (c *CredentialValue) MarshalJSON() ([]byte, error) {
	impl := credentialImpl{Version: 1}

	switch c.cvtype {
	case stringCV:
		impl.Body.Type = "string"
	case intCV:
		impl.Body.Type = "number"
	case floatCV:
		impl.Body.Type = "number"
	case unsetCV:
		impl.Body.Type = "undefined"
	case undecryptedCV:
		impl.Body.Type = "undecrypted"
	}

	if c.cvtype != unsetCV && c.cvtype != undecryptedCV {
		v, err := json.Marshal(c.raw)
		if err != nil {
			return nil, err
		}

		impl.Body.Value = v
	} else {
		impl.Body.Value = []byte(`""`)
	}

	b, err := json.Marshal(&impl)
	if err != nil {
		return nil, err
	}

	return []byte(strconv.Quote(string(b))), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *CredentialValue) UnmarshalJSON(b []byte) error {
	impl := credentialImpl{}

	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		v := reflect.ValueOf(c).Elem()
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	err = json.Unmarshal([]byte(s), &impl)
	if err != nil {
		return err
	}

	switch impl.Body.Type {
	case "undefined":
		c.cvtype = unsetCV
	case "undecrypted":
		c.cvtype = undecryptedCV
	case "string":
		c.cvtype = stringCV
		var v string
		err := json.Unmarshal(impl.Body.Value, &v)
		if err != nil {
			return errMistmatchedType
		}

		c.raw = v
		c.value = v
	case "number":
		c.cvtype = stringCV
		var v json.Number
		err := json.Unmarshal(impl.Body.Value, &v)
		if err != nil {
			return errMistmatchedType
		}

		if i, err := v.Int64(); err == nil {
			c.cvtype = intCV
			c.raw = i
		} else if f, err := v.Float64(); err == nil {
			c.cvtype = floatCV
			c.raw = f
		}

		c.value = v.String()
	default:
		return errors.New("Decoding type " + impl.Body.Type + " is not supported")
	}
	return nil
}

// NewUnsetCredentialValue creates a CredentialValue with an unset value.
func NewUnsetCredentialValue() *CredentialValue {
	return &CredentialValue{cvtype: unsetCV}
}

// NewStringCredentialValue creates a CredentialValue with a string value.
func NewStringCredentialValue(s string) *CredentialValue {
	return &CredentialValue{
		cvtype: stringCV,
		value:  s,
		raw:    s,
	}
}

// NewIntCredentialValue creates a CredentialValue with an int value.
func NewIntCredentialValue(i int) *CredentialValue {
	return &CredentialValue{
		cvtype: intCV,
		value:  strconv.Itoa(i),
		raw:    i,
	}
}

// NewFloatCredentialValue creates a CredentialValue with a float value.
func NewFloatCredentialValue(f float64) *CredentialValue {
	return &CredentialValue{
		cvtype: floatCV,
		value:  strconv.FormatFloat(f, 'g', -1, 64),
		raw:    f,
	}
}

// NewUndecryptedCredentialValue creates a CredentialValue with an undecrypted
// value
func NewUndecryptedCredentialValue() *CredentialValue {
	return &CredentialValue{
		cvtype: undecryptedCV,
	}
}
