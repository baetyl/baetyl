package server

import (
	"fmt"
	"net"
	"net/url"

	"github.com/baetyl/baetyl/master/database"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	context "golang.org/x/net/context"
)

// KVServer kv server to handle message
type KVServer struct {
	db database.DB
}

// NewKVServer creates a new kv server
func NewKVServer(log logger.Logger, db database.DB) error {
	uri := &url.URL{
		Scheme: "tcp",
		Host:   "127.0.0.1:50060",
	}
	listener, err := net.Listen(uri.Scheme, uri.Host)
	if err != nil {
		return err
	}
	log.Infoln("kv server is listening at: " + uri.Scheme + "://" + uri.Host)

	rpcServer := grpc.NewServer()
	baetyl.RegisterKVServer(rpcServer, &KVServer{db: db})
	reflection.Register(rpcServer)
	go func() {
		if err := rpcServer.Serve(listener); err != nil {
			log.Infoln("kv server shutdown: %v", err)
		}
		rpcServer.GracefulStop()
	}()
	return nil
}

// GetKV get kv
func (s *KVServer) GetKV(ctx context.Context, msg *baetyl.KVMessage) (*baetyl.KVMessage, error) {
	if len(msg.Key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	v, err := s.db.GetKV(msg.Key)
	if err != nil {
		return nil, err
	}
	return &baetyl.KVMessage{
		Key:   v.Key,
		Value: v.Value,
	}, nil
}

// ListKV list kvs with prefix
func (s *KVServer) ListKV(ctx context.Context, msg *baetyl.KVMessage) (*baetyl.ArrayKVMessage, error) {
	if len(msg.Key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	vs, err := s.db.ListKV(msg.Key)
	if err != nil {
		return nil, err
	}
	var kvs []*baetyl.KVMessage
	for _, kv := range vs {
		kvs = append(kvs, &baetyl.KVMessage{
			Key:   kv.Key,
			Value: kv.Value,
		})
	}
	return &baetyl.ArrayKVMessage{
		Kvs: kvs,
	}, nil
}

// PutKV put kv
func (s *KVServer) PutKV(ctx context.Context, msg *baetyl.KVMessage) (*baetyl.KVMessage, error) {
	if len(msg.Key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}
	if len(msg.Value) == 0 {
		return nil, fmt.Errorf("value is empty")
	}

	err := s.db.PutKV(msg.Key, msg.Value)
	return &baetyl.KVMessage{}, err
}

// DelKV del kv
func (s *KVServer) DelKV(ctx context.Context, msg *baetyl.KVMessage) (*baetyl.KVMessage, error) {
	if len(msg.Key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	err := s.db.DelKV(msg.Key)
	return &baetyl.KVMessage{}, err
}
