package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arigatomachine/cli/daemon/base64"
)

// PublicKey is en ed25519 public key.
type PublicKey struct {
	PublicKey base64.Value `json:"public_key"`
}

func loadPublicKey(prefs *preferences) (*PublicKey, error) {
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
