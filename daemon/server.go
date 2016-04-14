package main

import "errors"
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
	if len(socketPath) == 0 {
		return nil, errors.New("Provided socket path is empty")
	}

	absPath, err := filepath.Abs(socketPath)
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", absPath)
	if err != nil {
		return nil, err
	}

	return &Server{listener: l, socketPath: absPath, clientCount: 0}, nil
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
