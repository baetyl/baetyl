package api

import (
	"context"

	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
	"github.com/gogo/protobuf/types"
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
func (s *KVService) Set(_ context.Context, kv *api.KV) (*types.Empty, error) {
	return new(types.Empty), s.kv.Set(kv)
}

// Get get kv
func (s *KVService) Get(_ context.Context, kv *api.KV) (*api.KV, error) {
	return s.kv.Get(kv.Key)
}

// Del del kv
func (s *KVService) Del(_ context.Context, kv *api.KV) (*types.Empty, error) {
	return new(types.Empty), s.kv.Del(kv.Key)
}

// List list kvs with prefix
func (s *KVService) List(_ context.Context, kv *api.KV) (*api.KVs, error) {
	return s.kv.List(kv.Key)
}
