package base64

import (
	"encoding/base64"
	"errors"
	"reflect"
)

type Value []byte

func (bv *Value) MarshalJSON() ([]byte, error) {
	return []byte("\"" + base64.RawURLEncoding.EncodeToString(*bv) + "\""), nil
}

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
