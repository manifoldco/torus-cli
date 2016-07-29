package main

import (
	"fmt"
	"os"

	"github.com/nightlyone/lockfile"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/session"
	"github.com/arigatomachine/cli/daemon/socket"
)

type Daemon struct {
	proxy       *socket.AuthProxy
	lock        lockfile.Lockfile // actually a string
	session     session.Session
	config      *config.Config
	db          *db.DB
	hasShutdown bool
}

func NewDaemon(cfg *config.Config) (*Daemon, error) {

	lock, err := lockfile.New(cfg.PidPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create lockfile object: %s", err)
	}

	// Checks if there is a file; if there is an error and its not a
	// `isnotexists` error then return it back to the callee
	_, err = lock.GetOwner()
	if err != nil && !os.IsNotExist(err) &&
		err != lockfile.ErrInvalidPid && err != lockfile.ErrDeadOwner {
		return nil, err
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

	session := session.NewSession()

	db, err := db.NewDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	proxy, err := socket.NewAuthProxy(cfg, session, db)
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

func (d *Daemon) Addr() string {
	return d.proxy.Addr()
}

func (d *Daemon) Run() error {
	return d.proxy.Listen()
}

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
