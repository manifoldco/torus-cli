package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/natefinch/lumberjack"

	"github.com/arigatomachine/cli/daemon/config"
)

func main() {
	cfg, err := config.NewConfig(os.Getenv("ARIGATO_ROOT"))
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   path.Join(cfg.ArigatoRoot, "daemon.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	daemon, err := NewDaemon(cfg)
	if err != nil {
		log.Fatalf("Failed to create daemon: %s", err)
	}

	go watch(daemon)
	defer daemon.Shutdown()

	log.Printf("v%s of the Daemon is now listening on %s", cfg.Version, daemon.Addr())
	log.Printf("Daemon connecting to %s", c.API)
	err = daemon.Run()
	if err != nil {
		log.Printf("Error while running daemon: %s", err)
	}
}

func watch(daemon *Daemon) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c

	log.Printf("Caught a signal: %s", s)
	shutdown(daemon)
}

func shutdown(daemon *Daemon) {
	err := daemon.Shutdown()
	if err != nil {
		log.Printf("Did not shutdown cleanly, error: %s", err)
	}

	if r := recover(); r != nil {
		log.Printf("Failed shutting down; caught panic: %v", r)
		panic(r)
	}
}
