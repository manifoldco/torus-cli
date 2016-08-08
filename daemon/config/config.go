// Package config exposes static configuration data, and loaded user
// preferences.
package config

import (
	"fmt"
	"os"
	"path"
)

var version = "alpha"
var apiVersion = "0.1.0"

const requiredPermissions = 0700

// Config represents the static and user defined configuration data
// for Arigato.
type Config struct {
	ArigatoRoot string
	API         string
	APIVersion  string
	Version     string
	SocketPath  string
	PidPath     string
	DBPath      string
	PublicKey   *PublicKey
}

// NewConfig returns a new Config, with loaded user preferences.
func NewConfig(arigatoRoot string) (*Config, error) {
	if len(arigatoRoot) == 0 {
		arigatoRoot = path.Join(os.Getenv("HOME"), ".arigato")
	}

	src, err := os.Stat(arigatoRoot)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil && !src.IsDir() {
		return nil, fmt.Errorf("%s exists but is not a dir", arigatoRoot)
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(arigatoRoot, requiredPermissions)
		if err != nil {
			return nil, err
		}

		src, err = os.Stat(arigatoRoot)
		if err != nil {
			return nil, err
		}
	}

	fMode := src.Mode()
	if fMode.Perm() != requiredPermissions {
		return nil, fmt.Errorf("%s has permissions %d requires %d",
			arigatoRoot, fMode.Perm(), requiredPermissions)
	}

	prefs, err := newPreferences()
	if err != nil {
		return nil, err
	}

	publicKey, err := loadPublicKey(prefs)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ArigatoRoot: arigatoRoot,
		API:         prefs.Core.RegistryURI,
		APIVersion:  apiVersion,
		Version:     version,
		SocketPath:  path.Join(arigatoRoot, "daemon.socket"),
		PidPath:     path.Join(arigatoRoot, "daemon.pid"),
		DBPath:      path.Join(arigatoRoot, "daemon.db"),
		PublicKey:   publicKey,
	}

	return cfg, nil
}
