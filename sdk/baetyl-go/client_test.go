package baetyl

import (
	"context"
	"os"
	"testing"

	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvClient(t *testing.T) {
	got, err := NewEnvClient()
	assert.EqualError(t, err, "Env (BAETYL_MASTER_API_ADDRESS) not found")
	assert.Nil(t, got)

	server, err := api.NewServer(api.ServerConfig{
		Address: "tcp://127.0.0.1:50060",
	}, &mockMaster{})
	assert.NoError(t, err)
	defer server.Close()
	kvService := &mockKVService{
		m: make(map[string][]byte),
	}
	server.RegisterKVService(kvService)
	err = server.Start()
	assert.NoError(t, err)

	// old
	os.Setenv(EnvMasterAPIKey, "0.0.0.0")
	os.Setenv(EnvMasterAPIVersionKey, "v0")
	os.Setenv(EnvKeyMasterGRPCAPIAddress, "127.0.0.1:50060")
	got, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "/v0", got.ver)

	// new
	os.Setenv(EnvKeyMasterAPIAddress, "0.0.0.1")
	os.Setenv(EnvKeyMasterAPIVersion, "v1")
	os.Setenv(EnvKeyMasterGRPCAPIAddress, "127.0.0.1:50060")
	got, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "/v1", got.ver)

	_, err = got.GetKV([]byte("aa"))
	assert.NoError(t, err)

	err = got.SetKV([]byte("aa"), []byte("aadata"))
	assert.NoError(t, err)

	resp, err := got.GetKV([]byte("aa"))
	assert.NoError(t, err)
	assert.Equal(t, resp, []byte("aadata"))

	err = got.DelKV([]byte("aa"))
	assert.NoError(t, err)

	err = got.SetKV([]byte("a"), []byte("aa"))
	assert.NoError(t, err)

	err = got.SetKV([]byte("b"), []byte("bb"))
	assert.NoError(t, err)

	err = got.SetKV([]byte("c"), []byte("cc"))
	assert.NoError(t, err)

	respa, err := got.ListKV([]byte(""))
	assert.NoError(t, err)
	assert.Len(t, respa, 3)
}

type mockMaster struct{}

func (*mockMaster) Auth(u, p string) bool {
	return true
}

// KVService kv server
type mockKVService struct {
	m map[string][]byte
}

// Set set kv
func (s *mockKVService) Set(ctx context.Context, kv *api.KV) (*empty.Empty, error) {
	s.m[string(kv.Key)] = kv.Value
	return new(empty.Empty), nil
}

// Get get kv
func (s *mockKVService) Get(ctx context.Context, kv *api.KV) (*api.KV, error) {
	return &api.KV{
		Key:   kv.Key,
		Value: s.m[string(kv.Key)],
	}, nil
}

// Del del kv
func (s *mockKVService) Del(ctx context.Context, kv *api.KV) (*empty.Empty, error) {
	delete(s.m, string(kv.Key))
	return new(empty.Empty), nil
}

// List list kvs with prefix
func (s *mockKVService) List(ctx context.Context, kv *api.KV) (*api.KVs, error) {
	kvs := api.KVs{
		Kvs: []*api.KV{},
	}
	for k, v := range s.m {
		kvs.Kvs = append(kvs.Kvs, &api.KV{
			Key:   []byte(k),
			Value: v,
		})
	}
	return &kvs, nil
}
