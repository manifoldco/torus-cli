package socket

import (
	"net"
	"os"
	"path/filepath"
)

func MakeSocket(socketPath string) (net.Listener, error) {
	absPath, err := filepath.Abs(socketPath)
	if err != nil {
		return nil, err
	}

	// Attempt to remove an existing socket at this path if it exists.
	// Guarding against a server already running is outside the scope of this
	// module.
	err = os.Remove(absPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	l, err := net.Listen("unix", absPath)
	if err != nil {
		return nil, err
	}

	// Does not guarantee security; BSD ignores file permissions for sockets
	// see https://github.com/arigatomachine/cli/issues/76 for details
	if err = os.Chmod(socketPath, 0700); err != nil {
		return nil, err
	}

	return l, nil
}
