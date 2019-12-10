package api

import (
	"context"
	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
	"github.com/golang/protobuf/ptypes/empty"
)

// KV kv interface
type KV interface {
	Set(kv *api.KV) error
	Get(key []byte) (*api.KV, error)
	Del(key []byte) error
	List(prefix []byte) (*api.KVs, error)
}

// KVService kv server
type KVService struct {
	kv KV
}

// NewKVService new kv service
func NewKVService(kv KV) api.KVServiceServer {
	return &KVService{kv: kv}
}

// Set set kv
func (s *KVService) Set(ctx context.Context, kv *api.KV) (*empty.Empty, error) {
	return new(empty.Empty), s.kv.Set(kv)
}

// Get get kv
func (s *KVService) Get(ctx context.Context, kv *api.KV) (*api.KV, error) {
	return s.kv.Get(kv.Key)
}

// Del del kv
func (s *KVService) Del(ctx context.Context, kv *api.KV) (*empty.Empty, error) {
	return new(empty.Empty), s.kv.Del(kv.Key)
}

// List list kvs with prefix
func (s *KVService) List(ctx context.Context, kv *api.KV) (*api.KVs, error) {
	return s.kv.List(kv.Key)
}
