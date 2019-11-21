package api

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func Test_APIServer(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := database.New(database.Conf{Driver: "sqlite3", Source: path.Join(dir, "kv.db")})
	assert.NoError(t, err)
	defer db.Close()

	conf := Conf{Address: "baetyl"}
	apiServer := NewAPIServer(conf)
	assert.NotEmpty(t, apiServer)
	apiServer.RegisterKVService(NewKVService(db))
	err = apiServer.Start()
	assert.Error(t, err)
	if apiServer != nil {
		apiServer.Close()
	}

	conf = Conf{Address: "tcp://127.0.0.1:10000000"}
	apiServer = NewAPIServer(conf)
	assert.NotEmpty(t, apiServer)
	apiServer.RegisterKVService(NewKVService(db))
	err = apiServer.Start()
	assert.Error(t, err)
	if apiServer != nil {
		apiServer.Close()
	}

	ctx := context.Background()
	confs := []struct {
		server string
		client string
	}{
		{
			server: "tcp://127.0.0.1:50060",
			client: "127.0.0.1:50060",
		}, {
			server: "unix:///tmp/baetyl/api.sock",
			client: "unix:///tmp/baetyl/api.sock",
		},
	}
	for _, conf := range confs {
		apiServer = NewAPIServer(Conf{Address: conf.server})
		assert.NotEmpty(t, apiServer)
		apiServer.RegisterKVService(NewKVService(db))
		err = apiServer.Start()
		assert.NoError(t, err)

		conn, err := grpc.Dial(conf.client, grpc.WithInsecure())
		assert.NoError(t, err)
		client := baetyl.NewKVServiceClient(conn)
		assert.NotEmpty(t, client)

		_, err = client.Get(ctx, &baetyl.KV{Key: []byte("")})
		assert.NoError(t, err)

		resp, err := client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
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

		resp, err = client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
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
