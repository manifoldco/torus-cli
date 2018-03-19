// +build !windows

package prefs

import (
	"errors"
	"os"
	"path"
)

// ErrMissingHome represents an error where the $HOME dir could not be found
var ErrMissingHome = errors.New("Could not establish users home directory, is $HOME set?")

// RcPath returns the torusrc filepath
func RcPath() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", ErrMissingHome
	}

	return path.Join(home, rcFilename), nil
}
