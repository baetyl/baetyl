package api

import (
	"context"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/utils"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

func Test_APIServer(t *testing.T) {
	master := new(mockMaster)
	kvService := &mockKVService{
		m: make(map[string][]byte),
	}

	conf := ServerConfig{Address: "baetyl"}
	server, err := NewServer(conf, master)
	assert.NoError(t, err)
	assert.NotEmpty(t, server)
	server.RegisterKVService(kvService)
	err = server.Start()
	assert.Error(t, err)
	if server != nil {
		server.Close()
	}

	conf = ServerConfig{Address: "tcp://127.0.0.1:10000000"}
	server, err = NewServer(conf, master)
	assert.NoError(t, err)
	assert.NotEmpty(t, server)
	server.RegisterKVService(kvService)
	err = server.Start()
	assert.Error(t, err)
	if server != nil {
		server.Close()
	}

	server, err = NewServer(ServerConfig{Address: "tcp://127.0.0.1:50061"}, master)
	assert.NoError(t, err)
	assert.NotEmpty(t, server)
	server.RegisterKVService(kvService)
	err = server.Start()
	assert.NoError(t, err)

	cli, err := NewClient(ClientConfig{
		Address: "127.0.0.1:50061",
		Timeout: time.Duration(10) * time.Second,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, cli)

	_, err = cli.KV.Get(context.Background(), &KV{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unauthenticated desc = forbidden")

	server.Close()
	cli.Close()

	server, err = NewServer(ServerConfig{Address: "tcp://127.0.0.1:50062"}, master)
	assert.NoError(t, err)
	assert.NotEmpty(t, server)
	server.RegisterKVService(kvService)
	err = server.Start()
	assert.NoError(t, err)

	cli, err = NewClient(ClientConfig{
		Address:  "127.0.0.1:50062",
		Timeout:  time.Duration(10) * time.Second,
		Username: "baetyl",
		Password: "unknown",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, cli)

	_, err = cli.KV.Get(context.Background(), &KV{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unauthenticated desc = forbidden")

	server.Close()
	cli.Close()

	ctx := context.Background()
	confs := []struct {
		serverConf ServerConfig
		cliConf    ClientConfig
	}{
		{
			serverConf: ServerConfig{
				Address: "tcp://127.0.0.1:50060",
				Certificate: utils.Certificate{
					CA:   "./testcert/ca.pem",
					Key:  "./testcert/server.key",
					Cert: "./testcert/server.pem",
					Name: "bd",
				},
			},
			cliConf: ClientConfig{
				Address:  "127.0.0.1:50060",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
				Certificate: utils.Certificate{
					CA:   "./testcert/ca.pem",
					Key:  "./testcert/client.key",
					Cert: "./testcert/client.pem",
					Name: "bd",
				},
			},
		},
		{
			serverConf: ServerConfig{
				Address: "unix:///tmp/baetyl/api.sock",
				Certificate: utils.Certificate{
					CA:   "./testcert/ca.pem",
					Key:  "./testcert/server.key",
					Cert: "./testcert/server.pem",
					Name: "bd",
				},
			},
			cliConf: ClientConfig{
				Address:  "unix:///tmp/baetyl/api.sock",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
				Certificate: utils.Certificate{
					CA:   "./testcert/ca.pem",
					Key:  "./testcert/client.key",
					Cert: "./testcert/client.pem",
					Name: "bd",
				},
			},
		},
		{
			serverConf: ServerConfig{
				Address: "tcp://127.0.0.1:50060",
			},
			cliConf: ClientConfig{
				Address:  "127.0.0.1:50060",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
			},
		},
		{
			serverConf: ServerConfig{
				Address: "unix:///tmp/baetyl/api.sock",
			},
			cliConf: ClientConfig{
				Address:  "unix:///tmp/baetyl/api.sock",
				Timeout:  10 * time.Second,
				Username: "baetyl",
				Password: "baetyl",
			},
		},
	}
	for _, conf := range confs {
		server, err = NewServer(ServerConfig{Address: conf.serverConf.Address, Certificate: conf.serverConf.Certificate}, master)
		assert.NoError(t, err)
		assert.NotEmpty(t, server)
		kvService.m = make(map[string][]byte)
		server.RegisterKVService(kvService)
		err = server.Start()
		assert.NoError(t, err)

		cli, err = NewClient(conf.cliConf)
		assert.NoError(t, err)
		assert.NotEmpty(t, cli)

		_, err = cli.KV.Get(ctx, &KV{Key: []byte("aa")})
		assert.NoError(t, err)

		_, err = cli.KV.Set(ctx, &KV{Key: []byte("aa")})
		assert.NoError(t, err)

		_, err = cli.KV.Set(ctx, &KV{Key: []byte("aa"), Value: []byte("")})
		assert.NoError(t, err)

		_, err = cli.KV.Set(ctx, &KV{Key: []byte("aa"), Value: []byte("aadata")})
		assert.NoError(t, err)

		resp, err := cli.KV.Get(ctx, &KV{Key: []byte("aa")})
		assert.NoError(t, err)
		assert.Equal(t, resp.Key, []byte("aa"))
		assert.Equal(t, resp.Value, []byte("aadata"))

		_, err = cli.KV.Del(ctx, &KV{Key: []byte("aa")})
		assert.NoError(t, err)

		_, err = cli.KV.Del(ctx, &KV{Key: []byte("")})
		assert.NoError(t, err)

		resp, err = cli.KV.Get(ctx, &KV{Key: []byte("aa")})
		assert.NoError(t, err)
		assert.Equal(t, resp.Key, []byte("aa"))
		assert.Empty(t, resp.Value)

		_, err = cli.KV.Set(ctx, &KV{Key: []byte("/a"), Value: []byte("/root/a")})
		assert.NoError(t, err)

		_, err = cli.KV.Set(ctx, &KV{Key: []byte("/b"), Value: []byte("/root/b")})
		assert.NoError(t, err)

		_, err = cli.KV.Set(ctx, &KV{Key: []byte("/c"), Value: []byte("/root/c")})
		assert.NoError(t, err)

		respa, err := cli.KV.List(ctx, &KV{Key: []byte("/")})
		assert.NoError(t, err)
		assert.Len(t, respa.Kvs, 3)

		server.Close()
		cli.Close()
	}
}

type mockMaster struct{}

func (*mockMaster) Auth(u, p string) bool {
	if u == "baetyl" && p == "baetyl" {
		return true
	}
	return false
}

// KVService kv server
type mockKVService struct {
	m map[string][]byte
}

// Set set kv
func (s *mockKVService) Set(ctx context.Context, kv *KV) (*empty.Empty, error) {
	s.m[string(kv.Key)] = kv.Value
	return new(empty.Empty), nil
}

// Get get kv
func (s *mockKVService) Get(ctx context.Context, kv *KV) (*KV, error) {
	return &KV{
		Key:   kv.Key,
		Value: s.m[string(kv.Key)],
	}, nil
}

// Del del kv
func (s *mockKVService) Del(ctx context.Context, kv *KV) (*empty.Empty, error) {
	delete(s.m, string(kv.Key))
	return new(empty.Empty), nil
}

// List list kvs with prefix
func (s *mockKVService) List(ctx context.Context, kv *KV) (*KVs, error) {
	kvs := KVs{
		Kvs: []*KV{},
	}
	for k, v := range s.m {
		kvs.Kvs = append(kvs.Kvs, &KV{
			Key:   []byte(k),
			Value: v,
		})
	}
	return &kvs, nil
}
