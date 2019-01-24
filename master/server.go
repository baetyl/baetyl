package master

import (
	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/protocol/jrpc"
)

type master interface {
	reload(string) error
	stats() *openedge.Inspect
}

// Server for master API
type Server struct {
	s *jrpc.Server
	m master
}

func newServer(addr string, m master) (*Server, error) {
	s, err := jrpc.NewServer(addr)
	if err != nil {
		return nil, err
	}
	server := &Server{s, m}
	go server.s.Start("openedge", server)
	return server, nil
}

// InspectSystem inspect
func (s *Server) InspectSystem(args *openedge.InspectSystemRequest, reply *openedge.InspectSystemResponse) error {
	reply.Inspect = s.m.stats()
	return nil
}

// UpdateSystem reload
func (s *Server) UpdateSystem(args *openedge.UpdateSystemRequest, reply *openedge.UpdateSystemResponse) error {
	return s.m.reload(args.Config)
}

// // StartService method
// func (s *Server) StartService(args *openedge.StartServiceRequest, reply *openedge.StartServiceResponse) error {
// 	if s.m.services.Has(args.Name) {
// 		*reply = "duplicated"
// 		return nil
// 	}
// 	svc, err := s.m.engine.RunWithConfig(args.Name, &args.Info, args.Config)
// 	if err != nil {
// 		*reply = openedge.StartServiceResponse(err.Error())
// 	} else {
// 		*reply = ""
// 		s.m.services.Set(args.Name, svc)
// 	}
// 	return nil
// }
