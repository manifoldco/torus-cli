package daemon

import (
	"fmt"

	"github.com/nightlyone/lockfile"

	"github.com/manifoldco/torus-cli/config"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/db"
	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/registry"
	"github.com/manifoldco/torus-cli/daemon/session"
	"github.com/manifoldco/torus-cli/daemon/socket"
)

// Daemon is the torus coprocess that contains session secrets, handles
// cryptographic operations, and communication with the registry.
type Daemon struct {
	proxy       *socket.AuthProxy
	lock        lockfile.Lockfile // actually a string
	session     session.Session
	config      *config.Config
	db          *db.DB
	hasShutdown bool
}

// New creates a new Daemon.
func New(cfg *config.Config) (*Daemon, error) {
	lock, err := lockfile.New(cfg.PidPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create lockfile object: %s", err)
	}

	err = lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to create lockfile[%s]: %s", cfg.PidPath, err)
	}

	// Recover from the panic and return the error; this way we can
	// delete the lockfile!
	defer func() {
		if r := recover(); r != nil {
			err, _ = r.(error)
		}
	}()

	db, err := db.NewDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	session := session.NewSession()
	cryptoEngine := crypto.NewEngine(session)
	transport := socket.CreateHTTPTransport(cfg)
	client := registry.NewClient(cfg.RegistryURI.String(), cfg.APIVersion,
		cfg.Version, session, transport)
	logic := logic.NewEngine(cfg, session, db, cryptoEngine, client)

	proxy, err := socket.NewAuthProxy(cfg, session, db, transport, client, logic)
	if err != nil {
		return nil, fmt.Errorf("Failed to create auth proxy: %s", err)
	}

	daemon := &Daemon{
		proxy:       proxy,
		lock:        lock,
		session:     session,
		config:      cfg,
		db:          db,
		hasShutdown: false,
	}

	return daemon, nil
}

// Addr returns the domain socket the Daemon is listening on.
func (d *Daemon) Addr() string {
	return d.proxy.Addr()
}

// Run starts the daemon main loop. It returns on failure, or when the daemon
// has been gracefully shut down.
func (d *Daemon) Run() error {
	return d.proxy.Listen()
}

// Shutdown gracefully shuts down the daemon.
func (d *Daemon) Shutdown() error {
	if d.hasShutdown {
		return nil
	}

	d.hasShutdown = true
	if err := d.lock.Unlock(); err != nil {
		return fmt.Errorf("Could not unlock: %s", err)
	}

	if err := d.proxy.Close(); err != nil {
		return fmt.Errorf("Could not stop http proxy: %s", err)
	}

	if err := d.db.Close(); err != nil {
		return fmt.Errorf("Could not close db: %s", err)
	}

	return nil
}
