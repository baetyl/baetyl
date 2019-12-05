package api

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/master/engine"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Test_APIServer(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := database.New(database.Conf{Driver: "sqlite3", Source: path.Join(dir, "kv.db")})
	assert.NoError(t, err)
	defer db.Close()

	conf := Conf{Address: "baetyl"}
	apiServer, err := NewAPIServer(conf, &mockMaster{})
	assert.NoError(t, err)
	assert.NotEmpty(t, apiServer)
	apiServer.RegisterKVService(NewKVService(db))
	err = apiServer.Start()
	assert.Error(t, err)
	if apiServer != nil {
		apiServer.Close()
	}

	conf = Conf{Address: "tcp://127.0.0.1:10000000"}
	apiServer, err = NewAPIServer(conf, &mockMaster{})
	assert.NoError(t, err)
	assert.NotEmpty(t, apiServer)
	apiServer.RegisterKVService(NewKVService(db))
	err = apiServer.Start()
	assert.Error(t, err)
	if apiServer != nil {
		apiServer.Close()
	}

	ctx := context.Background()
	apiServer, err = NewAPIServer(Conf{Address: "tcp://127.0.0.1:50061"}, &mockMaster{})
	assert.NoError(t, err)
	assert.NotEmpty(t, apiServer)
	apiServer.RegisterKVService(NewKVService(db))
	err = apiServer.Start()
	assert.NoError(t, err)

	conn, err := grpc.Dial("127.0.0.1:50061", grpc.WithInsecure())
	assert.NoError(t, err)
	client := baetyl.NewKVServiceClient(conn)
	assert.NotEmpty(t, client)

	_, err = client.Get(ctx, &baetyl.KV{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unauthenticated desc = username or password not match")

	apiServer.Close()
	conn.Close()

	apiServer, err = NewAPIServer(Conf{Address: "tcp://127.0.0.1:50062"}, &mockMaster{})
	assert.NoError(t, err)
	assert.NotEmpty(t, apiServer)
	apiServer.RegisterKVService(NewKVService(db))
	err = apiServer.Start()
	assert.NoError(t, err)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithPerRPCCredentials(&customCred{
		Data: map[string]string{
			headerKeyUsername: "baetyl",
			headerKeyPassword: "unknown",
		},
	}))
	conn, err = grpc.Dial("127.0.0.1:50062", opts...)
	assert.NoError(t, err)
	client = baetyl.NewKVServiceClient(conn)
	assert.NotEmpty(t, client)

	_, err = client.Get(ctx, &baetyl.KV{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unauthenticated desc = username or password not match")

	apiServer.Close()
	conn.Close()

	confs := []struct {
		server     string
		client     string
		serverCert utils.Certificate
		clientCert utils.Certificate
	}{
		{
			server: "tcp://127.0.0.1:50060",
			client: "127.0.0.1:50060",
			serverCert: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
				CA:   "./testcert/ca.pem",
			},
			clientCert: utils.Certificate{
				Cert: "./testcert/client.pem",
				Key:  "./testcert/client.key",
				CA:   "./testcert/ca.pem",
			},
		},
		{
			server: "unix:///tmp/baetyl/api.sock",
			client: "unix:///tmp/baetyl/api.sock",
			serverCert: utils.Certificate{
				Cert: "./testcert/server.pem",
				Key:  "./testcert/server.key",
				CA:   "./testcert/ca.pem",
			},
			clientCert: utils.Certificate{
				Cert: "./testcert/client.pem",
				Key:  "./testcert/client.key",
				CA:   "./testcert/ca.pem",
			},
		},
		{
			server:     "tcp://127.0.0.1:50060",
			client:     "127.0.0.1:50060",
			serverCert: utils.Certificate{},
			clientCert: utils.Certificate{},
		},
		{
			server:     "unix:///tmp/baetyl/api.sock",
			client:     "unix:///tmp/baetyl/api.sock",
			serverCert: utils.Certificate{},
			clientCert: utils.Certificate{},
		},
	}
	for _, conf := range confs {
		apiServer, err = NewAPIServer(Conf{Address: conf.server, Certificate: conf.serverCert}, &mockMaster{})
		assert.NoError(t, err)
		assert.NotEmpty(t, apiServer)
		apiServer.RegisterKVService(NewKVService(db))
		err = apiServer.Start()
		assert.NoError(t, err)

		var opts []grpc.DialOption
		if conf.clientCert.Cert == "" {
			opts = append(opts, grpc.WithInsecure())
		} else {
			tlsCfg, err := utils.NewTLSClientConfig(conf.clientCert)
			assert.NoError(t, err)
			if tlsCfg != nil {
				tlsCfg.InsecureSkipVerify = true
				creds := credentials.NewTLS(tlsCfg)
				opts = append(opts, grpc.WithTransportCredentials(creds))
			}
		}
		opts = append(opts, grpc.WithPerRPCCredentials(&customCred{
			Data: map[string]string{
				headerKeyUsername: "baetyl",
				headerKeyPassword: "baetyl",
			},
		}))
		conn, err := grpc.Dial(conf.client, opts...)
		assert.NoError(t, err)
		client := baetyl.NewKVServiceClient(conn)
		assert.NotEmpty(t, client)

		_, err = client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
		assert.NoError(t, err)

		_, err = client.Set(ctx, &baetyl.KV{})
		assert.Error(t, err)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("")})
		assert.Error(t, err)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("aa")})
		assert.NoError(t, err)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("aa"), Value: []byte("")})
		assert.NoError(t, err)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("aa"), Value: []byte("aadata")})
		assert.NoError(t, err)

		resp, err := client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
		assert.NoError(t, err)
		assert.Equal(t, resp.Key, []byte("aa"))
		assert.Equal(t, resp.Value, []byte("aadata"))

		_, err = client.Del(ctx, &baetyl.KV{Key: []byte("aa")})
		assert.NoError(t, err)

		_, err = client.Del(ctx, &baetyl.KV{Key: []byte("")})
		assert.NoError(t, err)

		resp, err = client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
		assert.NoError(t, err)
		assert.Equal(t, resp.Key, []byte("aa"))
		assert.Empty(t, resp.Value)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("/root/a"), Value: []byte("/root/ax")})
		assert.NoError(t, err)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("/root/b"), Value: []byte("/root/bx")})
		assert.NoError(t, err)

		_, err = client.Set(ctx, &baetyl.KV{Key: []byte("/roox/a"), Value: []byte("/roox/ax")})
		assert.NoError(t, err)

		respa, err := client.List(ctx, &baetyl.KV{Key: []byte("/root")})
		assert.NoError(t, err)
		assert.Len(t, respa.Kvs, 2)
		assert.Equal(t, respa.Kvs[0].Key, []byte("/root/a"))
		assert.Equal(t, respa.Kvs[1].Key, []byte("/root/b"))
		assert.Equal(t, respa.Kvs[0].Value, []byte("/root/ax"))
		assert.Equal(t, respa.Kvs[1].Value, []byte("/root/bx"))

		respa, err = client.List(ctx, &baetyl.KV{Key: []byte("/roox")})
		assert.NoError(t, err)
		assert.Len(t, respa.Kvs, 1)
		assert.Equal(t, respa.Kvs[0].Key, []byte("/roox/a"))
		assert.Equal(t, respa.Kvs[0].Value, []byte("/roox/ax"))

		_, err = client.Del(ctx, &baetyl.KV{Key: []byte("/root/a")})
		assert.NoError(t, err)

		_, err = client.Del(ctx, &baetyl.KV{Key: []byte("/root/b")})
		assert.NoError(t, err)

		_, err = client.Del(ctx, &baetyl.KV{Key: []byte("/roox/a")})
		assert.NoError(t, err)

		apiServer.Close()
		conn.Close()
	}
}

type mockMaster struct{}

func (*mockMaster) Auth(u, p string) bool {
	if u == "baetyl" && p == "baetyl" {
		return true
	}
	return false
}

func (*mockMaster) InspectSystem() ([]byte, error) {
	return nil, nil
}

func (*mockMaster) UpdateSystem(trace, tp, target string) error {
	return nil
}

func (*mockMaster) ReportInstance(serviceName, instanceName string, partialStats engine.PartialStats) error {
	return nil
}

func (*mockMaster) StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	return nil
}

func (*mockMaster) StopInstance(serviceName, instanceName string) error {
	return nil
}

type customCred struct {
	Data map[string]string
}

// GetRequestMetadata & RequireTransportSecurity for Custom Credential
// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (c *customCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return c.Data, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (c *customCred) RequireTransportSecurity() bool {
	return false
}
