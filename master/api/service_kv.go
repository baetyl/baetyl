package api

import (
	"context"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/golang/protobuf/ptypes/empty"
)

// KVService kv server
type KVService struct {
	m Master
}

// NewKVService new kv service
func NewKVService(m Master) *KVService {
	return &KVService{m: m}
}

// Set set kv
func (s *KVService) Set(ctx context.Context, kv *baetyl.KV) (*empty.Empty, error) {
	return new(empty.Empty), s.m.SetKV(kv)
}

// Get get kv
func (s *KVService) Get(ctx context.Context, kv *baetyl.KV) (*baetyl.KV, error) {
	return s.m.GetKV(kv.Key)
}

// Del del kv
func (s *KVService) Del(ctx context.Context, kv *baetyl.KV) (*empty.Empty, error) {
	return new(empty.Empty), s.m.DelKV(kv.Key)
}

// List list kvs with prefix
func (s *KVService) List(ctx context.Context, kv *baetyl.KV) (*baetyl.KVs, error) {
	return s.m.ListKV(kv.Key)
}

// RegisterKVService register kv service
func (s *APIServer) RegisterKVService(server baetyl.KVServiceServer) {
	baetyl.RegisterKVServiceServer(s.svr, server)
}
