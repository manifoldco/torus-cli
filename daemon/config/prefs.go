package config

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/go-ini/ini"
	"github.com/kardianos/osext"
)

const (
	rcFilename        = ".arigatorc"
	publicKeyFilename = "public_key.json"
	caBundleFilename  = "ca_bundle.pem"
	registryURI       = "https://registry.arigato.sh"
)

type preferences struct {
	Core core `ini:"core"`
}

type core struct {
	PublicKeyFile string `ini:"public_key_file"`
	CABundleFile  string `ini:"ca_bundle_file"`
	RegistryURI   string `ini:"registry_uri"`
}

func newPreferences() (*preferences, error) {
	dir, err := osext.ExecutableFolder()
	if err != nil {
		return nil, fmt.Errorf("Could not determine executable location: %s", err)
	}

	// certs and keys live in the root of the node package
	defaultKeyPath := path.Join(dir, "..", publicKeyFilename)
	defaultBundlePath := path.Join(dir, "..", caBundleFilename)
	prefs := &preferences{
		Core: core{
			PublicKeyFile: defaultKeyPath,
			CABundleFile:  defaultBundlePath,
			RegistryURI:   registryURI,
		},
	}

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(u.HomeDir, rcFilename)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return prefs, nil
	}

	if err != nil {
		return nil, err
	}

	err = ini.MapTo(prefs, filePath)
	if err != nil {
		return nil, err
	}

	return prefs, nil
}
