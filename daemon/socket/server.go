package socket

import "os"
import "net"
import "path/filepath"

type Server struct {
	clientCount int
	listener    net.Listener
	socketPath  string
}

type Listener interface {
	Accept() (Client, error)
	Close() error
	Addr() net.Addr
	String() string
}

func NewServer(socketPath string) (*Server, error) {

	absPath, err := filepath.Abs(socketPath)
	if err != nil {
		return nil, err
	}

	// Stat the socket path to determine if it exists or not. Guarding against
	// a server already running is outside the scope of this module.
	_, err = os.Stat(socketPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// If file exist then we can remove it!
	if !os.IsNotExist(err) {
		err = os.Remove(socketPath)
		if err != nil {
			return nil, err
		}
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

	return &Server{listener: l, socketPath: socketPath, clientCount: 0}, nil
}

func (s *Server) Accept() (Client, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}

	s.clientCount += 1
	return NewConnection(conn, s.clientCount), nil
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) String() string {
	return s.listener.Addr().String()
}
