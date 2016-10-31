// Package base64 provides a byte slice type that marshals into json as
// a raw (no padding) base64url value.
package base64

import (
	"encoding/base64"
	"errors"
	"reflect"
)

// Value is a base64url encoded json object,
type Value []byte

// NewValue returns a pointer to a Value, cast from the given byte slice.
// This is a convenience function, handling the address creation that a direct
// cast would not allow.
func NewValue(b []byte) *Value {
	v := Value(b)
	return &v
}

// NewValueFromString returns a Value containing the decoded data in encoded.
func NewValueFromString(encoded string) (*Value, error) {
	out, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return NewValue(out), nil
}

// MarshalJSON returns the ba64url encoding of bv for JSON representation.
func (bv *Value) MarshalJSON() ([]byte, error) {
	return []byte("\"" + base64.RawURLEncoding.EncodeToString(*bv) + "\""), nil
}

func (bv *Value) String() string {
	return base64.RawURLEncoding.EncodeToString(*bv)
}

// UnmarshalJSON sets bv to the bytes represented in the base64url encoding b.
func (bv *Value) UnmarshalJSON(b []byte) error {
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
