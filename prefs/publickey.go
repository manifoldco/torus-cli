package prefs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/base64"
)

const requiredPermissions = 0700

// PublicKey is en ed25519 public key.
type PublicKey struct {
	PublicKey base64.Value `json:"public_key"`
}

// LoadPublicKey reads the publickey file from disk and parses the json
func LoadPublicKey(prefs *Preferences) (*PublicKey, error) {
	filePath := prefs.Core.PublicKeyFile

	fd, err := readPublicKeyFile(filePath)
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

func parsePublicKeyFile(fd *os.File) (*PublicKey, error) {
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
		text = "error: must be a file that exists"
		return cli.NewExitError(text, -1)
	}

	fMode := src.Mode()
	if fMode.Perm() != requiredPermissions {
		text = fmt.Sprintf("error: file specified has permissions %d, must have permissions %d", fMode.Perm(), requiredPermissions)
		return cli.NewExitError(text, -1)
	}

	fd, err := readPublicKeyFile(filePath)
	if err != nil {
		text = "error: could not read file, permissions ok"
		return cli.NewExitError(text, -1)
	}

	_, err = parsePublicKeyFile(fd)
	if err != nil {
		text = "error: could not parse json"
		return cli.NewExitError(text, -1)
	}

	return nil
}
