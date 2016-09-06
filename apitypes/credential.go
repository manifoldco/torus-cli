package apitypes

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/arigatomachine/cli/identity"
)

var errMistmatchedType = errors.New("Mismatched type and value in credential")

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
	PathExp   string           `json:"pathexp"`
	ProjectID *identity.ID     `json:"project_id"`
	Value     *CredentialValue `json:"value"`
}

// CredentialValue is the raw value of a credential.
type CredentialValue struct {
	unset bool
	value string
}

// IsUnset returns if this credential has been unset (deleted)
func (c *CredentialValue) IsUnset() bool {
	return c.unset
}

// String returns the string representation of this credential. It panics
// if the credential was deleted.
func (c *CredentialValue) String() string {
	if c.unset {
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

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *CredentialValue) UnmarshalJSON(b []byte) error {
	c.unset = false

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
		c.unset = true
	case "string":
		var v string
		err := json.Unmarshal(impl.Body.Value, &v)
		if err != nil {
			return errMistmatchedType
		}

		c.value = v
	case "number":
		var v json.Number
		err := json.Unmarshal(impl.Body.Value, &v)
		if err != nil {
			return errMistmatchedType
		}

		c.value = v.String()
	default:
		return errors.New("Decoding type " + impl.Body.Type + " is not supported")
	}
	return nil
}
