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
	m   *Master
	l   net.Listener
	log openedge.Logger
}

func newServer(m *Master) (*Server, error) {
	addr, err := url.Parse(m.cfg.Server)
	if err != nil {
		return nil, err
	}
	l, err := net.Listen(addr.Scheme, addr.Host)
	if err != nil {
		return nil, err
	}

	s := &Server{m, l, openedge.WithField("openedge", "server")}
	srv := rpc.NewServer()
	err = srv.RegisterName("openedge", s)
	if err != nil {
		l.Close()
		return nil, err
	}
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				s.log.Debugln(err.Error())
				return
			}
			go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	return s, nil
}

func (s *Server) close() {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			s.log.Warnln(err.Error())
		}
	}
}

// UpdateSystem reload
func (s *Server) UpdateSystem(args *openedge.UpdateSystemRequest, reply *openedge.UpdateSystemResponse) error {
	return s.m.reload(args.Config)
}

// StartService method
func (s *Server) StartService(args *openedge.StartServiceRequest, reply *openedge.StartServiceResponse) error {
	if s.m.services.Has(args.Name) {
		*reply = "duplicated"
		return nil
	}
	svc, err := s.m.engine.RunWithConfig(args.Name, &args.Info, args.Config)
	if err != nil {
		*reply = openedge.StartServiceResponse(err.Error())
	} else {
		*reply = ""
		s.m.services.Set(args.Name, svc)
	}
	return nil
}
