package api

import (
	"net"
	"syscall"

	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	context "golang.org/x/net/context"
)

// Conf the configuration of database
type Conf struct {
	Address string
}

// APIServer api server to handle grpc message
type APIServer struct {
	db  database.DB
	svr *grpc.Server
}

// NewAPIServer creates a new api server
func NewAPIServer(conf Conf, db database.DB, log logger.Logger) (*APIServer, error) {
	uri, err := utils.ParseURL(conf.Address)
	if err != nil {
		return nil, err
	}

	if uri.Scheme == "unix" {
		if err := syscall.Unlink(uri.Host); err != nil {
			log.Errorf(err.Error())
		}
	}

	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return nil, err
	}
	log.Infof("api server is listening at: %s", uri.String())

	svr := grpc.NewServer()
	apiServer := &APIServer{db: db, svr: svr}
	baetyl.RegisterKVServiceServer(svr, apiServer)
	reflection.Register(svr)
	go func() {
		if err := svr.Serve(listener); err != nil {
			log.Infof("api server shutdown: %v", err)
		}
	}()
	return apiServer, nil
}

// Set set kv
func (s *APIServer) Set(ctx context.Context, kv *baetyl.KV) (*baetyl.KV, error) {
	return &baetyl.KV{}, s.db.Set(kv)
}

// Get get kv
func (s *APIServer) Get(ctx context.Context, kv *baetyl.KV) (*baetyl.KV, error) {
	return s.db.Get(kv.Key)
}

// Del del kv
func (s *APIServer) Del(ctx context.Context, msg *baetyl.KV) (*baetyl.KV, error) {
	return &baetyl.KV{}, s.db.Del(msg.Key)
}

// List list kvs with prefix
func (s *APIServer) List(ctx context.Context, kv *baetyl.KV) (*baetyl.KVs, error) {
	return s.db.List(kv.Key)
}

// Close closes server
func (s *APIServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
