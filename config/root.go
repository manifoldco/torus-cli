// +build !windows

package config

import (
	"os"
	"path"
)

const requiredPermissions = 0700

func torusRootPath() string {
	torusRoot := os.Getenv("TORUS_ROOT")
	if len(torusRoot) == 0 {
		torusRoot = path.Join(os.Getenv("HOME"), ".torus")
	}

	return torusRoot
}
