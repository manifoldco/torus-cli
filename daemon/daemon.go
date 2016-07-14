package main

import (
	"fmt"
	"os"

	"github.com/nightlyone/lockfile"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/session"
	"github.com/arigatomachine/cli/daemon/socket"
)

type Daemon struct {
	server      socket.Listener
	proxy       *socket.AuthProxy
	lock        lockfile.Lockfile // actually a string
	session     session.Session
	config      *config.Config
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

	server, err := socket.NewServer(cfg.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct server: %s", err)
	}

	session := session.NewSession()

	proxy, err := socket.NewAuthProxy(cfg, session)
	if err != nil {
		return nil, fmt.Errorf("Failed to create auth proxy: %s", err)
	}

	daemon := &Daemon{
		server:      server,
		proxy:       proxy,
		lock:        lock,
		session:     session,
		config:      cfg,
		hasShutdown: false,
	}

	return daemon, nil
}

func (d *Daemon) Run() error {
	d.proxy.Listen()

	for {
		client, err := d.server.Accept()
		if err != nil {
			return err
		}

		router := NewRouter(client, d.config, d.session)
		go router.process()
	}
}

func (d *Daemon) Shutdown() error {
	if d.hasShutdown {
		return nil
	}

	d.hasShutdown = true
	if err := d.lock.Unlock(); err != nil {
		return fmt.Errorf("Could not unlock: %s", err)
	}

	if err := d.server.Close(); err != nil {
		return fmt.Errorf("Could not shutdown server: %s", err)
	}

	if err := d.proxy.Close(); err != nil {
		return fmt.Errorf("Could not stop http proxy: %s", err)
	}

	return nil
}
