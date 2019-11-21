package api

import (
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Master master interface
type Master interface {
	Auth(u, p string) bool

	// for system
	InspectSystem() ([]byte, error)
	UpdateSystem(trace, tp, target string) error

	// for instance
	ReportInstance(serviceName, instanceName string, partialStats engine.PartialStats) error
	StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
	StopInstance(serviceName, instanceName string) error

	// for db
	SetKV(kv *baetyl.KV) error
	GetKV(key []byte) (*baetyl.KV, error)
	DelKV(key []byte) error
	ListKV(prefix []byte) (*baetyl.KVs, error)

	Logger() logger.Logger
}

// Conf the configuration of database
type Conf struct {
	Address string
}

// APIServer api server to handle grpc message
type APIServer struct {
	m    Master
	conf Conf
	svr  *grpc.Server
}

// NewAPIServer creates a new api server
func NewAPIServer(conf Conf, m Master) (*APIServer, error) {
	svr := grpc.NewServer()
	apiServer := &APIServer{m: m, conf: conf, svr: svr}
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
			s.m.Logger().Errorf(err.Error())
		}
		dir := filepath.Dir(uri.Host)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			s.m.Logger().Errorf("failed to make directory %s : %s", dir, err.Error())
		}
	}
	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return err
	}
	s.m.Logger().Infof("api server is listening at: %s://%s", uri.Scheme, uri.Host)
	reflection.Register(s.svr)
	go func() {
		if err := s.svr.Serve(listener); err != nil {
			s.m.Logger().Infof("api server shutdown: %v", err)
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
