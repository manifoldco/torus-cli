package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

type Base64Value []byte

func (bv *Base64Value) MarshalJSON() ([]byte, error) {
	return []byte("\"" + base64.RawURLEncoding.EncodeToString(*bv) + "\""), nil
}

func (bv *Base64Value) UnmarshalJSON(b []byte) error {
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

type PublicKey struct {
	PublicKey Base64Value `json:"public_key"`
}

func loadPublicKey(prefs *Preferences) (*PublicKey, error) {
	filePath := prefs.Core.PublicKeyFile

	fd, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Could not locate public key file: %s", filePath)
	}

	if err != nil {
		return nil, err
	}

	key := &PublicKey{}
	dec := json.NewDecoder(fd)

	err = dec.Decode(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
