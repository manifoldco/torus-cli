package cmd

// Listener is an interface for daemon processes that listen on some socket.
type Listener interface {
	// Addr returns the address of the running service. This is a socket in
	// the case of a daemon, or an TCP port in the case of the Gateway.
	Addr() string

	// Run starts the Service. This operation will block.
	Run() error

	// Shutdown shuts the Service, and does any cleanup
	Shutdown() error
}

