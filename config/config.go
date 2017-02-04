// Package config exposes static configuration data, and loaded user
// preferences.
package config

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/manifoldco/torus-cli/data"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
)

// Version is the compiled version of our binary. It is set via the Makefile.
var Version = "alpha"
var apiVersion = "0.4.0"

// Config represents the static and user defined configuration data
// for Torus.
type Config struct {
	APIVersion string
	Version    string

	TorusRoot      string
	SocketPath     string
	PidPath        string
	DBPath         string
	LastUpdatePath string

	RegistryURI *url.URL
	CABundle    *x509.CertPool
	PublicKey   *prefs.PublicKey
}

// NewConfig returns a new Config, with loaded user preferences.
func NewConfig(torusRoot string) (*Config, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return nil, err
	}

	publicKey, err := prefs.LoadPublicKey(preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key")
	}

	caBundle, err := loadCABundle(preferences.Core.CABundleFile)
	if err != nil {
		return nil, err
	}

	registryURI, err := url.Parse(preferences.Core.RegistryURI)
	if err != nil {
		return nil, fmt.Errorf("invalid registry_uri")
	}

	cfg := &Config{
		APIVersion: apiVersion,
		Version:    Version,

		TorusRoot:      torusRoot,
		SocketPath:     path.Join(torusRoot, "daemon.socket"),
		PidPath:        path.Join(torusRoot, "daemon.pid"),
		DBPath:         path.Join(torusRoot, "daemon.db"),
		LastUpdatePath: path.Join(torusRoot, "last_update"),

		RegistryURI: registryURI,
		CABundle:    caBundle,
		PublicKey:   publicKey,
	}

	return cfg, nil
}

// CreateTorusRoot creates the root directory for the Torus daemon.
func CreateTorusRoot(checkPermissions bool) (string, error) {
	torusRoot := torusRootPath()
	src, err := os.Stat(torusRoot)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err == nil && !src.IsDir() {
		return "", fmt.Errorf("%s exists but is not a dir", torusRoot)
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(torusRoot, requiredPermissions)
		if err != nil {
			return "", err
		}

		src, err = os.Stat(torusRoot)
		if err != nil {
			return "", err
		}
	}

	fMode := src.Mode()
	if checkPermissions && fMode.Perm() != requiredPermissions {
		return "", fmt.Errorf("%s has permissions %d requires %d",
			torusRoot, fMode.Perm(), requiredPermissions)
	}

	return torusRoot, nil
}

// Load CABundle creates a new CertPool from the given filename
func loadCABundle(cafile string) (*x509.CertPool, error) {
	var pem []byte
	var err error

	if cafile == "" {
		pem, err = data.Asset("data/ca_bundle.pem")
	} else {
		pem, err = ioutil.ReadFile(cafile)

	}
	if err != nil {
		return nil, fmt.Errorf("unable to find CA bundle")
	}

	c := x509.NewCertPool()
	ok := c.AppendCertsFromPEM(pem)
	if !ok {
		return nil, fmt.Errorf("unable to load CA bundle from %s", cafile)
	}

	return c, nil
}

// LoadConfig loads the config, standardizing cli errors on failure.
func LoadConfig() (*Config, error) {
	cfg, err := NewConfig(torusRootPath())
	if err != nil {
		return nil, errs.NewErrorExitError("Failed to load config.", err)
	}

	return cfg, nil
}
