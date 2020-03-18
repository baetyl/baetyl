package baetyl

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvClient(t *testing.T) {
	cli, err := NewEnvClient()
	assert.EqualError(t, err, "Env (BAETYL_MASTER_API_ADDRESS) not found")
	assert.Nil(t, cli)

	server, err := apiserver.NewServer(apiserver.ServerConfig{
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
		serverConf apiserver.ServerConfig
		cliConf    apiserver.ClientConfig
	}{
		{
			serverConf: apiserver.ServerConfig{
				Address: "tcp://127.0.0.1:51060",
			},
			cliConf: apiserver.ClientConfig{
				Address:  "127.0.0.1:51060",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
			},
		},
		{
			serverConf: apiserver.ServerConfig{
				Address: "unix:///tmp/baetyl/run/apiserver.sock",
			},
			cliConf: apiserver.ClientConfig{
				Address:  "unix:///tmp/baetyl/run/apiserver.sock",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
			},
		},
	}
	for _, conf := range confs {
		server, err := apiserver.NewServer(apiserver.ServerConfig{Address: conf.serverConf.Address, Certificate: conf.serverConf.Certificate}, master)
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

		a := apiserver.KV{
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

		ctx0, cel0 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel0()
		err = cli.SetKVConext(ctx0, a)
		assert.NoError(t, err)

		server.Close()

		ctx1, cel1 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel1()
		_, err = cli.GetKVConext(ctx1, a.Key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		ctx2, cel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel2()
		err = cli.SetKVConext(ctx2, a)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		ctx3, cel3 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel3()
		resp, err = cli.GetKVConext(ctx3, a.Key)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		ctx4, cel4 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel4()
		err = cli.DelKVConext(ctx4, a.Key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		ctx5, cel5 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel5()
		respa, err = cli.ListKVContext(ctx5, []byte(""))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadlineExceeded desc")

		server, err = apiserver.NewServer(apiserver.ServerConfig{Address: conf.serverConf.Address, Certificate: conf.serverConf.Certificate}, master)
		assert.NoError(t, err)
		assert.NotEmpty(t, server)
		kvService.m = make(map[string][]byte)
		server.RegisterKVService(kvService)
		err = server.Start()
		assert.NoError(t, err)

		a.Key = []byte("aa")
		err = cli.SetKV(a)
		assert.NoError(t, err)

		ctx6, cel6 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel6()
		resp, err = cli.GetKVConext(ctx6, a.Key)
		assert.NoError(t, err)
		assert.Equal(t, resp.Value, a.Value)

		ctx7, cel7 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel7()
		err = cli.DelKVConext(ctx7, a.Key)
		assert.NoError(t, err)

		ctx8, cel8 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel8()
		err = cli.SetKVConext(ctx8, a)
		assert.NoError(t, err)

		ctx9, cel9 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel9()
		a.Key = []byte("bb")
		err = cli.SetKVConext(ctx9, a)
		assert.NoError(t, err)

		ctx10, cel10 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cel10()
		respa, err = cli.ListKVContext(ctx10, []byte(""))
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
func (s *mockKVService) Set(ctx context.Context, kv *apiserver.KV) (*types.Empty, error) {
	s.m[string(kv.Key)] = kv.Value
	return new(types.Empty), nil
}

// Get get kv
func (s *mockKVService) Get(ctx context.Context, kv *apiserver.KV) (*apiserver.KV, error) {
	return &apiserver.KV{
		Key:   kv.Key,
		Value: s.m[string(kv.Key)],
	}, nil
}

// Del del kv
func (s *mockKVService) Del(ctx context.Context, kv *apiserver.KV) (*types.Empty, error) {
	delete(s.m, string(kv.Key))
	return new(types.Empty), nil
}

// List list kvs with prefix
func (s *mockKVService) List(ctx context.Context, kv *apiserver.KV) (*apiserver.KVs, error) {
	kvs := apiserver.KVs{
		Kvs: []*apiserver.KV{},
	}
	for k, v := range s.m {
		kvs.Kvs = append(kvs.Kvs, &apiserver.KV{
			Key:   []byte(k),
			Value: v,
		})
	}
	return &kvs, nil
}
