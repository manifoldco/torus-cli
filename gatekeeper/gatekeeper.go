// Package gatekeeper is a web service that will listen for a machine identity from a Cloud
// provider, and request machine credentials.
package gatekeeper

import (
	"fmt"
	"os"

	"github.com/nightlyone/lockfile"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/gatekeeper/http"
)

// Gatekeeper provides the Listener interface for a daemonized HTTP co-process that will act
// to register machines
type Gatekeeper struct {
	lock lockfile.Lockfile
	g    *http.Gatekeeper
}

// New returns a new Gatekeeper
func New(cfg *config.Config, groupShared bool) (g *Gatekeeper, err error) {
	lock, err := lockfile.New(cfg.GatekeeperPidPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create lockfile object: %s", err)
	}

	err = lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("Failed to create lockfile[%s]: %s", cfg.GatekeeperPidPath, err)
	}

	// recovery and cleanup
	defer func() {
		if r := recover(); r != nil {
			// named return
			g = nil
			err = r.(error)
		}
	}()

	if groupShared {
		if err := os.Chmod(string(lock), 0600); err != nil {
			return nil, err
		}
	}

	api := api.NewClient(cfg)
	http := http.NewGatekeeper(cfg, api)

	gatekeeper := &Gatekeeper{
		lock: lock,
		g:    http,
	}

	return gatekeeper, nil
}

// Addr returns the address of the currently running Gatekeeper
func (g *Gatekeeper) Addr() string {
	return g.g.Addr()
}

// Run starts the Gatekeeper main loop. If blocks until the service is shutdown, or encounters an
// error
func (g *Gatekeeper) Run() error {
	return g.g.Listen()
}

// Shutdown gracefully stops the HTTP service
func (g *Gatekeeper) Shutdown() error {
	if err := g.lock.Unlock(); err != nil {
		return fmt.Errorf("Could not unlock: %s", err)
	}

	if err := g.g.Close(); err != nil {
		return fmt.Errorf("Could not stop gatekeeper: %s", err)
	}

	return nil
}
