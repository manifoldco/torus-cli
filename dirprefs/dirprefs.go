// Package dirprefs provides directory/project specific preference settings
package dirprefs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DirPreferences holds preferences for arguments set in .arigato.json files
type DirPreferences struct {
	Organization string `json:"org,omitempty"`
	Project      string `json:"project,omitempty"`
}

// Load loads DirPreferences. It starts in the current working directory,
// looking for a readable '.arigato.json' file, and walks up the directory
// hierarchy until it finds one, or reaches the root of the fs.
//
// It returns an empty DirPreferences is no '.arigato.json' files are found.
// It returns an error if a malformed file is found, or if any errors occur
// during file system access.
func Load() (*DirPreferences, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	prefs := &DirPreferences{}

	var f *os.File
	for {
		f, err = os.Open(filepath.Join(path, ".arigato.json"))
		if err != nil {
			if len(path) == 1 && path == string(os.PathSeparator) {
				return prefs, nil
			}

			path = filepath.Dir(path)
			continue
		}

		break
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(prefs)
	if err != nil {
		return nil, err
	}

	return prefs, nil
}
