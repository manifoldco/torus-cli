package socket

import "net"

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

	l, err := MakeSocket(socketPath)
	if err != nil {
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
