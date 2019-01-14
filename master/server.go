package master

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"

	openedge "github.com/baidu/openedge/api/go"
)

// Server for master API
type Server struct {
	m *Master
	l net.Listener
}

func (s *Server) start(m *Master) error {
	srv := rpc.NewServer()
	err := srv.RegisterName("openedge", s)
	if err != nil {
		return err
	}
	addr, err := url.Parse(m.cfg.Server)
	if err != nil {
		return err
	}
	s.l, err = net.Listen(addr.Scheme, addr.Host)
	if err != nil {
		return err
	}
	s.m = m
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				return
			}
			go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	return nil
}

func (s *Server) stop() {
	s.l.Close()
}

// StartService method
func (s *Server) StartService(args openedge.StartServiceRequest, reply *openedge.StartServiceResponse) error {
	if s.m.svcs.Has(args.Name) {
		*reply = "duplicated"
		return nil
	}
	svc, err := s.m.engine.RunWithConfig(args.Name, &args.Info, args.Config)
	if err != nil {
		*reply = openedge.StartServiceResponse(err.Error())
	} else {
		*reply = ""
		s.m.svcs.Set(args.Name, svc)
	}
	return nil
}
