package openedge

import (
	fmt "fmt"
	"net"

	"github.com/baidu/openedge/utils"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	// TODO: to test
	tls, err := utils.NewTLSServerConfig(c.Certificate)
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(c.Concurrent.Max),
		grpc.MaxRecvMsgSize(int(c.Message.Length.Max)),
		grpc.MaxSendMsgSize(int(c.Message.Length.Max)),
	}
	if tls != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tls)))
	}
	svr := grpc.NewServer(opts...)
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
