package jrpc

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/baidu/openedge/utils"
)

// Server json rpc server
type Server struct {
	s *rpc.Server
	l net.Listener
}

// NewServer creates a new json rpc server
func NewServer(addr string) (*Server, error) {
	url, err := utils.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	lis, err := net.Listen(url.Scheme, url.Host)
	if err != nil {
		return nil, err
	}
	srv := rpc.NewServer()
	return &Server{
		s: srv,
		l: lis,
	}, nil
}

// Start starts json rpc server
func (s *Server) Start(name string, handler interface{}) error {
	err := s.s.RegisterName(name, handler)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				return
			}
			go s.s.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	return nil
}

// Close closese server
func (s *Server) Close() error {
	return s.l.Close()
}

// Addr returns address
func (s *Server) Addr() string {
	return s.l.Addr().String()
}
