package main

import "os"
import "log"
import "os/signal"

var version = "development"

func main() {

	path := os.Getenv("SOCKET_PATH")
	server, err := NewServer(path)
	session := NewSession()

	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}

	defer shutdown(server)
	go watch(server)

	log.Printf("Listening on %s", server.Addr())
	for {
		client, err := server.Accept()
		if err != nil {
			log.Fatalf("Accept Error: %s", err)
		}

		router := NewRouter(client, session)
		go router.process()
	}
}

func watch(server *Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	s := <-c

	log.Printf("Caught a signal: %s", s)
	shutdown(server)
}

func shutdown(server *Server) {
	log.Printf("Shutting down server")

	if err := server.Close(); err != nil {
		log.Fatalf("Could not shutdown server: %s", err)
	}

	if r := recover(); r != nil {
		log.Printf("Failed shutting down; caught panic: %v", r)
		panic(r)
	}
}
