// Package dirprefs provides directory/project specific preference settings
package dirprefs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DirPreferences holds preferences for arguments set in .torus.json files
type DirPreferences struct {
	Organization string `json:"org,omitempty"`
	Project      string `json:"project,omitempty"`
	Path         string `json:"-"`
}

// Load loads DirPreferences. It starts in the current working directory,
// looking for a readable '.torus.json' file, and walks up the directory
// hierarchy until it finds one, or reaches the root of the fs.
//
// It returns an empty DirPreferences is no '.torus.json' files are found.
// It returns an error if a malformed file is found, or if any errors occur
// during file system access.
func Load(recurse bool) (*DirPreferences, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	prefs := &DirPreferences{}

	var f *os.File
	for {
		f, err = os.Open(filepath.Join(path, ".torus.json"))
		if err != nil {
			if len(path) == 1 && path == string(os.PathSeparator) || !recurse {
				return prefs, nil
			}

			path = filepath.Dir(path)
			continue
		}

		break
	}

	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(prefs)
	if err != nil {
		return nil, err
	}

	prefs.Path = f.Name()
	return prefs, nil
}

// Save writes the DirPreferences values to the file in the struct's Path
// field
func (d *DirPreferences) Save() error {
	f, err := os.Create(d.Path)
	if err != nil {
		return err
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(d)
}

// Remove removes the backing file for this DirPreferences
func (d *DirPreferences) Remove() error {
	return os.Remove(d.Path)
}
