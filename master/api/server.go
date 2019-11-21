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
	"github.com/golang/protobuf/ptypes/empty"
	context "golang.org/x/net/context"
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
	m   Master
	svr *grpc.Server
}

// NewAPIServer creates a new api server
func NewAPIServer(conf Conf, m Master) (*APIServer, error) {
	uri, err := utils.ParseURL(conf.Address)
	if err != nil {
		return nil, err
	}

	if uri.Scheme == "unix" {
		if err := syscall.Unlink(uri.Host); err != nil {
			m.Logger().Errorf(err.Error())
		}
		dir := filepath.Dir(uri.Host)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			m.Logger().Errorf("failed to make directory %s : %s", dir, err.Error())
		}
	}

	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return nil, err
	}
	m.Logger().Infof("api server is listening at: %s://%s", uri.Scheme, uri.Host)

	svr := grpc.NewServer()
	apiServer := &APIServer{m: m, svr: svr}
	baetyl.RegisterKVServiceServer(svr, apiServer)
	reflection.Register(svr)
	go func() {
		if err := svr.Serve(listener); err != nil {
			m.Logger().Infof("api server shutdown: %v", err)
		}
	}()
	return apiServer, nil
}

// Set set kv
func (s *APIServer) Set(ctx context.Context, kv *baetyl.KV) (*empty.Empty, error) {
	return new(empty.Empty), s.m.SetKV(kv)
}

// Get get kv
func (s *APIServer) Get(ctx context.Context, kv *baetyl.KV) (*baetyl.KV, error) {
	return s.m.GetKV(kv.Key)
}

// Del del kv
func (s *APIServer) Del(ctx context.Context, kv *baetyl.KV) (*empty.Empty, error) {
	return new(empty.Empty), s.m.DelKV(kv.Key)
}

// List list kvs with prefix
func (s *APIServer) List(ctx context.Context, kv *baetyl.KV) (*baetyl.KVs, error) {
	return s.m.ListKV(kv.Key)
}

// Close closes server
func (s *APIServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
