package config

import (
	"fmt"
	"os"
	"path"

	"github.com/go-ini/ini"
	"github.com/kardianos/osext"
	"github.com/mitchellh/go-homedir"
)

const (
	publicKeyFilename = "public_key.json"
	rcFilename        = ".arigatorc"
)

type preferences struct {
	Core core `ini:"core"`
}

type core struct {
	PublicKeyFile string `ini:"public_key_file"`
}

func newPreferences() (*preferences, error) {
	exeFolder, err := osext.ExecutableFolder()
	if err != nil {
		return nil, fmt.Errorf("Could not determine executable folder: %s", err)
	}

	defaultKeyPath := path.Join(exeFolder, publicKeyFilename)
	prefs := &preferences{
		Core: core{
			PublicKeyFile: defaultKeyPath,
		},
	}

	homePath, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(homePath, rcFilename)
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
