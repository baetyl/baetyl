package baetyl

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvClient(t *testing.T) {
	cli, err := NewEnvClient()
	assert.EqualError(t, err, "Env (BAETYL_MASTER_API_ADDRESS) not found")
	assert.Nil(t, cli)

	server, err := api.NewServer(api.ServerConfig{
		Address: "tcp://127.0.0.1:52060",
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
	os.Setenv(EnvKeyMasterGRPCAPIAddress, "127.0.0.1:52060")
	cli, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, cli)
	assert.Equal(t, "/v0", cli.ver)

	master := new(mockMaster)
	confs := []struct {
		serverConf api.ServerConfig
		cliConf    api.ClientConfig
	}{
		{
			serverConf: api.ServerConfig{
				Address: "tcp://127.0.0.1:51060",
			},
			cliConf: api.ClientConfig{
				Address:  "127.0.0.1:51060",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
			},
		},
		{
			serverConf: api.ServerConfig{
				Address: "unix:///tmp/baetyl/run/api.sock",
			},
			cliConf: api.ClientConfig{
				Address:  "unix:///tmp/baetyl/run/api.sock",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
			},
		},
	}
	for _, conf := range confs {
		server, err := api.NewServer(api.ServerConfig{Address: conf.serverConf.Address, Certificate: conf.serverConf.Certificate}, master)
		assert.NoError(t, err)
		assert.NotEmpty(t, server)
		kvService.m = make(map[string][]byte)
		server.RegisterKVService(kvService)
		err = server.Start()
		assert.NoError(t, err)

		// new
		os.Setenv(EnvKeyMasterAPIAddress, "0.0.0.1")
		os.Setenv(EnvKeyMasterAPIVersion, "v1")
		os.Setenv(EnvKeyMasterGRPCAPIAddress, conf.cliConf.Address)
		cli, err := NewEnvClient()
		assert.NoError(t, err)
		assert.NotNil(t, cli)
		assert.Equal(t, "/v1", cli.ver)

		a := api.KV{
			Key:   []byte("name"),
			Value: []byte("baetyl"),
		}
		_, err = cli.GetKV(a.Key)
		assert.NoError(t, err)

		err = cli.SetKV(a)
		assert.NoError(t, err)

		resp, err := cli.GetKV(a.Key)
		assert.NoError(t, err)
		assert.Equal(t, resp.Value, a.Value)

		err = cli.DelKV(a.Key)
		assert.NoError(t, err)

		err = cli.SetKV(a)
		assert.NoError(t, err)

		a.Key = []byte("bb")
		err = cli.SetKV(a)
		assert.NoError(t, err)

		respa, err := cli.ListKV([]byte(""))
		assert.NoError(t, err)
		assert.Len(t, respa, 2)

		server.Close()

		ctx, cel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cel()

		_, err = cli.GetKVConext(ctx, a.Key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		err = cli.SetKVConext(ctx, a)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		resp, err = cli.GetKVConext(ctx, a.Key)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		err = cli.DelKVConext(ctx, a.Key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		err = cli.SetKVConext(ctx, a)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		a.Key = []byte("bb")
		err = cli.SetKVConext(ctx, a)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		respa, err = cli.ListKVContext(ctx, []byte(""))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		server, err = api.NewServer(api.ServerConfig{Address: conf.serverConf.Address, Certificate: conf.serverConf.Certificate}, master)
		assert.NoError(t, err)
		assert.NotEmpty(t, server)
		kvService.m = make(map[string][]byte)
		server.RegisterKVService(kvService)
		err = server.Start()
		assert.NoError(t, err)

		_, err = cli.GetKV(a.Key)
		assert.NoError(t, err)

		a.Key = []byte("aa")
		err = cli.SetKV(a)
		assert.NoError(t, err)

		resp, err = cli.GetKV(a.Key)
		assert.NoError(t, err)
		assert.Equal(t, resp.Value, a.Value)

		err = cli.DelKV(a.Key)
		assert.NoError(t, err)

		err = cli.SetKV(a)
		assert.NoError(t, err)

		a.Key = []byte("bb")
		err = cli.SetKV(a)
		assert.NoError(t, err)

		respa, err = cli.ListKV([]byte(""))
		assert.NoError(t, err)
		assert.Len(t, respa, 2)

		server.Close()
		cli.Close()
	}
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
func (s *mockKVService) Set(ctx context.Context, kv *api.KV) (*types.Empty, error) {
	s.m[string(kv.Key)] = kv.Value
	return new(types.Empty), nil
}

// Get get kv
func (s *mockKVService) Get(ctx context.Context, kv *api.KV) (*api.KV, error) {
	return &api.KV{
		Key:   kv.Key,
		Value: s.m[string(kv.Key)],
	}, nil
}

// Del del kv
func (s *mockKVService) Del(ctx context.Context, kv *api.KV) (*types.Empty, error) {
	delete(s.m, string(kv.Key))
	return new(types.Empty), nil
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
