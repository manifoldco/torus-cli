package config

import (
	"fmt"
	"os"
	"path"

	"github.com/go-ini/ini"
	"github.com/kardianos/osext"
	"github.com/mitchellh/go-homedir"
)

const PUBLIC_KEY_FILENAME = "public_key.json"
const RC_FILENAME = ".arigatorc"

type Preferences struct {
	*Default
	*Core
}

type Default struct{}
type Core struct {
	PublicKeyFile string `ini:"public_key_file"`
}

func NewPreferences() (*Preferences, error) {
	exeFolder, err := osext.ExecutableFolder()
	if err != nil {
		return nil, fmt.Errorf("Could not determine executable folder: %s", err)
	}

	defaultKeyPath := path.Join(exeFolder, PUBLIC_KEY_FILENAME)
	prefs := &Preferences{
		Default: &Default{},
		Core: &Core{
			PublicKeyFile: defaultKeyPath,
		},
	}

	homePath, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(homePath, RC_FILENAME)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return prefs, nil
	}

	if err != nil {
		return nil, err
	}

	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, err
	}

	publicKeyFileValue := cfg.Section("core").Key("public_key_file").String()
	if publicKeyFileValue != "" {
		prefs.Core.PublicKeyFile = publicKeyFileValue
	}

	return prefs, nil
}
