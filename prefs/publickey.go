package prefs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/data"
	"github.com/manifoldco/torus-cli/errs"
)

const requiredPermissions = 0700

// PublicKey is en ed25519 public key.
type PublicKey struct {
	PublicKey base64.Value `json:"public_key"`
}

// LoadPublicKey reads the publickey file from disk and parses the json
func LoadPublicKey(prefs *Preferences) (*PublicKey, error) {
	filePath := prefs.Core.PublicKeyFile

	var fd io.Reader
	var err error

	if filePath == "" {
		var b []byte
		b, err = data.Asset("data/public_key.json")

		if err == nil {
			fd = bytes.NewReader(b)
		}
	} else {
		fd, err = readPublicKeyFile(filePath)
	}

	if err != nil {
		return nil, err
	}

	key, err := parsePublicKeyFile(fd)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func readPublicKeyFile(filePath string) (*os.File, error) {
	fd, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("error: could not locate public key file: %s", filePath)
	}
	if err != nil {
		return nil, err
	}

	return fd, nil
}

func parsePublicKeyFile(fd io.Reader) (*PublicKey, error) {
	key := &PublicKey{}
	dec := json.NewDecoder(fd)

	err := dec.Decode(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ValidatePublicKey checks the publickey path for valid file
func ValidatePublicKey(filePath string) error {
	var text string

	src, err := os.Stat(filePath)
	if err != nil {
		return errs.NewExitError("Publick key file must exist")
	}

	fMode := src.Mode()
	if fMode.Perm() != requiredPermissions {
		text = fmt.Sprintf("File specified has permissions %d, must have permissions %d", fMode.Perm(), requiredPermissions)
		return errs.NewExitError(text)
	}

	fd, err := readPublicKeyFile(filePath)
	if err != nil {
		return errs.NewExitError("Could not read file, permissions ok")
	}

	_, err = parsePublicKeyFile(fd)
	if err != nil {
		return errs.NewExitError("Could not parse JSON")
	}

	return nil
}
