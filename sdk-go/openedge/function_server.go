package openedge

import (
	fmt "fmt"
	"net"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Call message handler
type Call func(context.Context, *FunctionMessage) (*FunctionMessage, error)

// FServer functions server to handle message
type FServer struct {
	addr string
	cfg  FunctionServerConfig
	svr  *grpc.Server
	call Call
}

// NewFServer creates a new functions server
func NewFServer(c FunctionServerConfig, call Call) (*FServer, error) {
	lis, err := net.Listen("tcp", c.Address)
	if err != nil {
		return nil, err
	}
	svr := grpc.NewServer(
		// grpc.ConnectionTimeout(c.Timeout),
		grpc.MaxRecvMsgSize(int(c.Message.Length.Max)),
		grpc.MaxSendMsgSize(int(c.Message.Length.Max)),
	)
	s := &FServer{cfg: c, call: call, svr: svr, addr: lis.Addr().String()}
	RegisterFunctionServer(svr, s)
	reflection.Register(svr)
	go s.svr.Serve(lis)
	return s, nil
}

// Call handles message
func (s *FServer) Call(c context.Context, m *FunctionMessage) (*FunctionMessage, error) {
	if s.call == nil {
		return nil, fmt.Errorf("handle not implemented")
	}
	return s.call(c, m)
}

// Close closes server
func (s *FServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
