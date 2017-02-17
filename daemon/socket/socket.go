// +build !windows

package socket

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func makeSocket(transportAddress string, groupShared bool) (net.Listener, error) {
	absPath, err := filepath.Abs(transportAddress)
	if err != nil {
		fmt.Printf("Error getting absolute path: %s\n", transportAddress)
		return nil, err
	}

	// Attempt to remove an existing socket at this path if it exists.
	// Guarding against a server already running is outside the scope of this
	// module.
	err = os.Remove(absPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing absPath: %s\n", absPath)
		return nil, err
	}

	l, err := net.Listen("unix", absPath)
	if err != nil {
		fmt.Printf("Error listening on unix socket: %s\n", absPath)
		return nil, err
	}

	mode := os.FileMode(0700)
	if groupShared {
		mode = 0760
	}

	// Does not guarantee security; BSD ignores file permissions for sockets
	// see https://github.com/manifoldco/torus-cli/issues/76 for details
	if err = os.Chmod(transportAddress, mode); err != nil {
		fmt.Printf("Error chmodding socket: %s:%s\n", transportAddress, mode)
		return nil, err
	}

	return l, nil
}
