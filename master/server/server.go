package server

import (
	"net"

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
	Address string `yaml:"address" json:"address"`
}

// KVServer kv server to handle message
type KVServer struct {
	db  database.DB
	svr *grpc.Server
}

// NewKVServer creates a new kv server
func NewKVServer(conf Conf, db database.DB, log logger.Logger) (*KVServer, error) {
	uri, err := utils.ParseURL(conf.Address)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return nil, err
	}
	log.Infof("kv server is listening at: %s://%s", uri.Scheme, uri.Host)

	svr := grpc.NewServer()
	kvServer := &KVServer{db: db, svr: svr}
	baetyl.RegisterKVServiceServer(svr, kvServer)
	reflection.Register(svr)
	go func() {
		if err := svr.Serve(listener); err != nil {
			log.Infof("kv server shutdown: %v", err)
		}
	}()
	return kvServer, nil
}

// Set put kv
func (s *KVServer) Set(ctx context.Context, kv *baetyl.KV) (*baetyl.KV, error) {
	return &baetyl.KV{}, s.db.Set(kv)
}

// Get get kv
func (s *KVServer) Get(ctx context.Context, kv *baetyl.KV) (*baetyl.KV, error) {
	return s.db.Get(kv.Key)
}

// Del del kv
func (s *KVServer) Del(ctx context.Context, msg *baetyl.KV) (*baetyl.KV, error) {
	return &baetyl.KV{}, s.db.Del(msg.Key)
}

// List list kvs with prefix
func (s *KVServer) List(ctx context.Context, kv *baetyl.KV) (*baetyl.KVs, error) {
	return s.db.List(kv.Key)
}

// Close closes server
func (s *KVServer) Close() {
	if s.svr != nil {
		s.svr.GracefulStop()
	}
}
