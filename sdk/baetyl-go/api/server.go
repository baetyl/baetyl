package api

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/baetyl/baetyl-go/utils"
	"github.com/baetyl/baetyl/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Master interface
type Master interface {
	Auth(u, p string) bool
}

// NewServer creates a new server
func NewServer(conf ServerConfig, m Master) (*Server, error) {
	var opts []grpc.ServerOption
	if conf.Key != "" || conf.Cert != "" {
		tlsCfg, err := utils.NewTLSConfigServer(conf.Certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.Creds(creds))
	} else {
		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return false, status.Errorf(codes.Unauthenticated, "missing authentication information")
			}
			var u, p string
			if val, ok := md[headerKeyUsername]; ok {
				u = val[0]
			}
			if val, ok := md[headerKeyPassword]; ok {
				p = val[0]
			}
			ok = m.Auth(u, p)
			if !ok {
				return false, status.Errorf(codes.Unauthenticated, "forbidden")
			}
			return handler(ctx, req)
		}
		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}
	return &Server{conf: conf, svr: grpc.NewServer(opts...)}, nil
}

// RegisterKVService register kv service
func (s *Server) RegisterKVService(server KVServiceServer) {
	RegisterKVServiceServer(s.svr, server)
}

// Start start api server
func (s *Server) Start() error {
	uri, err := utils.ParseURL(s.conf.Address)
	if err != nil {
		return err
	}

	if uri.Scheme == "unix" {
		if err := syscall.Unlink(uri.Host); err != nil {
			logger.Errorf(err.Error())
		}
		dir := filepath.Dir(uri.Host)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			logger.Errorf("failed to make directory %s : %s", dir, err.Error())
		}
	}
	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return err
	}
	logger.Infof("api server is listening at: %s", s.conf.Address)
	reflection.Register(s.svr)
	go func() {
		if err := s.svr.Serve(listener); err != nil {
			logger.Infof("api server shutdown: %v", err)
		}
	}()
	return nil
}

// Close closes api server
func (s *Server) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
