package api

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	headerKeyUsername = "x-baetyl-username"
	headerKeyPassword = "x-baetyl-password"
)

// Conf the configuration of database
type Conf struct {
	Address           string `yaml:"address" json:"address"`
	utils.Certificate `yaml:",inline" json:",inline"`
}

// APIServer api server to handle grpc message
type APIServer struct {
	conf Conf
	svr  *grpc.Server
}

// NewAPIServer creates a new api server
func NewAPIServer(conf Conf, m Master) (*APIServer, error) {
	var opts []grpc.ServerOption
	tlsCfg, err := utils.NewTLSServerConfig(conf.Certificate)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.Creds(creds))
	}
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return false, status.Errorf(codes.Unauthenticated, "no metadata")
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
			return false, status.Errorf(codes.Unauthenticated, "username or password not match")
		}
		return handler(ctx, req)
	}
	opts = append(opts, grpc.UnaryInterceptor(interceptor))
	return &APIServer{conf: conf, svr: grpc.NewServer(opts...)}, nil
}

// Start start api server
func (s *APIServer) Start() error {
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
func (s *APIServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
