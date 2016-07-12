package main

import "os"
import "fmt"
import "errors"
import "github.com/nightlyone/lockfile"
import "github.com/arigatomachine/cli/daemon/socket"

type Daemon struct {
	server      socket.Listener
	lock        lockfile.Lockfile // actually a string
	session     Session
	config      *Config
	hasShutdown bool
}

func NewDaemon(arigatoRoot string) (*Daemon, error) {
	cfg, err := NewConfig(arigatoRoot)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to start: %s", err))
	}

	lock, err := lockfile.New(cfg.PidPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(
			"Failed to create lockfile object: %s", err))
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
		return nil, errors.New(fmt.Sprintf(
			"Failed to create lockfile[%s]: %s", cfg.PidPath, err))
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
		panic(errors.New(fmt.Sprintf("Failed to construct server: %s", err)))
	}

	session := NewSession()
	daemon := &Daemon{
		server:      server,
		lock:        lock,
		session:     session,
		config:      cfg,
		hasShutdown: false,
	}

	return daemon, nil
}

func (d *Daemon) Run() error {
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
		return errors.New(fmt.Sprintf("Could not unlock: %s", err))
	}

	if err := d.server.Close(); err != nil {
		return errors.New(fmt.Sprintf("Could not shutdown server: %s", err))
	}

	return nil
}
