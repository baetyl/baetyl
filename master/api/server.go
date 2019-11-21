package api

import (
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Conf the configuration of database
type Conf struct {
	Address string
}

// APIServer api server to handle grpc message
type APIServer struct {
	conf Conf
	svr  *grpc.Server
}

// NewAPIServer creates a new api server
func NewAPIServer(conf Conf) (*APIServer, error) {
	svr := grpc.NewServer()
	apiServer := &APIServer{conf: conf, svr: svr}
	return apiServer, nil
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
