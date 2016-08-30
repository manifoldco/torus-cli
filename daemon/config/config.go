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

	"github.com/arigatomachine/cli/prefs"
)

// Version is the compiled version of our binary. It is set via the Makefile.
var Version = "alpha"
var apiVersion = "0.1.0"

const requiredPermissions = 0700

// Config represents the static and user defined configuration data
// for Arigato.
type Config struct {
	APIVersion string
	Version    string

	ArigatoRoot string
	SocketPath  string
	PidPath     string
	DBPath      string

	RegistryURI *url.URL
	CABundle    *x509.CertPool
	PublicKey   *prefs.PublicKey
}

// NewConfig returns a new Config, with loaded user preferences.
func NewConfig(arigatoRoot string) (*Config, error) {
	preferences, err := prefs.NewPreferences(true)
	if err != nil {
		return nil, err
	}

	publicKey, err := prefs.LoadPublicKey(preferences)
	if err != nil {
		return nil, err
	}

	caBundle, err := loadCABundle(preferences.Core.CABundleFile)
	if err != nil {
		return nil, err
	}

	registryURI, err := url.Parse(preferences.Core.RegistryURI)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		APIVersion: apiVersion,
		Version:    Version,

		ArigatoRoot: arigatoRoot,
		SocketPath:  path.Join(arigatoRoot, "daemon.socket"),
		PidPath:     path.Join(arigatoRoot, "daemon.pid"),
		DBPath:      path.Join(arigatoRoot, "daemon.db"),

		RegistryURI: registryURI,
		CABundle:    caBundle,
		PublicKey:   publicKey,
	}

	return cfg, nil
}

// CreateArigatoRoot creates the root directory for the Arigato daemon.
func CreateArigatoRoot(arigatoRoot string) (string, error) {
	if len(arigatoRoot) == 0 {
		arigatoRoot = path.Join(os.Getenv("HOME"), ".arigato")
	}

	src, err := os.Stat(arigatoRoot)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err == nil && !src.IsDir() {
		return "", fmt.Errorf("%s exists but is not a dir", arigatoRoot)
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(arigatoRoot, requiredPermissions)
		if err != nil {
			return "", err
		}

		src, err = os.Stat(arigatoRoot)
		if err != nil {
			return "", err
		}
	}

	fMode := src.Mode()
	if fMode.Perm() != requiredPermissions {
		return "", fmt.Errorf("%s has permissions %d requires %d",
			arigatoRoot, fMode.Perm(), requiredPermissions)
	}

	return arigatoRoot, nil
}

// Load CABundle creates a new CertPool from the given filename
func loadCABundle(cafile string) (*x509.CertPool, error) {

	pem, err := ioutil.ReadFile(cafile)
	if err != nil {
		return nil, err
	}

	c := x509.NewCertPool()
	ok := c.AppendCertsFromPEM(pem)
	if !ok {
		return nil, fmt.Errorf("Unable to load CA bundle from %s", cafile)
	}

	return c, nil
}
