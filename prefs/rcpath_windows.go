package prefs

import (
	"errors"
	"os"
	"path"
)

// ErrMissingHome represents an error where the $HOMEDIR and/or $HOMEDRIVE
var ErrMissingHome = errors.New("Could not establish users home directory, is %HOMEDRIVE and %HOME set?")

// RcPath returns the torusrc filepath
func RcPath() (string, error) {
	homedrive := os.Getenv("HOMEDRIVE")
	if homedrive == "" {
		return "", ErrMissingHome
	}

	homedir := os.Getenv("HOMEDIR")
	if homedir == "" {
		return "", ErrMissingHome
	}

	return path.Join(homedrive, homedir, rcFilename), nil
}
