package apitypes

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/pathexp"
)

var errMistmatchedType = errors.New("Mismatched type and value in credential")

const (
	unsetCV = iota
	stringCV
	intCV
	floatCV
)

// CredentialEnvelope is an unencrypted credential object with a
// deserialized body
type CredentialEnvelope struct {
	ID      *identity.ID `json:"id"`
	Version uint8        `json:"version"`
	Body    *Credential  `json:"body"`
}

// Credential is the body of an unencrypted Credential
type Credential struct {
	Name      string           `json:"name"`
	OrgID     *identity.ID     `json:"org_id"`
	PathExp   *pathexp.PathExp `json:"pathexp"`
	ProjectID *identity.ID     `json:"project_id"`
	Value     *CredentialValue `json:"value"`
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

// String returns the string representation of this credential. It panics
// if the credential was deleted.
func (c *CredentialValue) String() string {
	if c.cvtype == unsetCV {
		panic("CredentialValue has been unset")
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
	}

	if c.cvtype != unsetCV {
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

	err = json.Unmarshal([]byte(s), &impl)
	if err != nil {
		return err
	}

	switch impl.Body.Type {
	case "undefined":
		c.cvtype = unsetCV
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
