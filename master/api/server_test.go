package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"google.golang.org/grpc"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func Test_APIServer(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := database.New(database.Conf{Driver: "sqlite3", Source: path.Join(dir, "kv.db")})
	assert.NoError(t, err)
	defer db.Close()

	log := newMockLogger()
	conf := Conf{Address: "baetyl"}
	_, err = NewAPIServer(conf, db, log)
	assert.Error(t, err)

	conf = Conf{Address: "tcp://127.0.0.1:10000000"}
	apiServer, err := NewAPIServer(conf, db, log)
	assert.Error(t, err)

	conf = Conf{Address: "tcp://127.0.0.1:50060"}
	apiServer, err = NewAPIServer(conf, db, log)
	assert.NoError(t, err)
	assert.Equal(t, []string{fmt.Sprintf("[Infof]api server is listening at: %s", conf.Address)}, log.records)
	defer apiServer.Close()

	conn, err := grpc.Dial("127.0.0.1:50060", grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()
	client := baetyl.NewKVServiceClient(conn)
	assert.NotEmpty(t, client)

	ctx := context.Background()
	_, err = client.Get(ctx, &baetyl.KV{Key: []byte("")})
	assert.NoError(t, err)

	resp, err := client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Equal(t, resp.Key, []byte("aa"))
	assert.Empty(t, resp.Value)

	resp, err = client.Set(ctx, &baetyl.KV{})
	assert.Error(t, err)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("")})
	assert.Error(t, err)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("aa"), Value: []byte("")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("aa"), Value: []byte("aadata")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	resp, err = client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Equal(t, resp.Key, []byte("aa"))
	assert.Equal(t, resp.Value, []byte("aadata"))

	resp, err = client.Del(ctx, &baetyl.KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	resp, err = client.Del(ctx, &baetyl.KV{Key: []byte("")})
	assert.NoError(t, err)

	resp, err = client.Get(ctx, &baetyl.KV{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Equal(t, resp.Key, []byte("aa"))
	assert.Empty(t, resp.Value)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("/root/a"), Value: []byte("/root/ax")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("/root/b"), Value: []byte("/root/bx")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	resp, err = client.Set(ctx, &baetyl.KV{Key: []byte("/roox/a"), Value: []byte("/roox/ax")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

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
}

type mackLogger struct {
	records []string
	data    map[string]interface{}
	err     error
}

func newMockLogger() *mackLogger {
	return &mackLogger{
		data:    map[string]interface{}{},
		records: []string{},
	}
}

func (l *mackLogger) WithField(key string, value interface{}) logger.Logger {
	l.data[key] = value
	return l
}
func (l *mackLogger) WithError(err error) logger.Logger {
	l.err = err
	return l
}
func (l *mackLogger) Debugf(format string, args ...interface{}) {
	l.records = append(l.records, "[Debugf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Infof(format string, args ...interface{}) {
	l.records = append(l.records, "[Infof]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Warnf(format string, args ...interface{}) {
	l.records = append(l.records, "[Warnf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Errorf(format string, args ...interface{}) {
	l.records = append(l.records, "[Errorf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Fatalf(format string, args ...interface{}) {
	l.records = append(l.records, "[Fatalf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Debugln(args ...interface{}) {
	l.records = append(l.records, "[Debugln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Infoln(args ...interface{}) {
	l.records = append(l.records, "[Infoln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Warnln(args ...interface{}) {
	l.records = append(l.records, "[Warnln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Errorln(args ...interface{}) {
	l.records = append(l.records, "[Errorln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Fatalln(args ...interface{}) {
	l.records = append(l.records, "[Fatalln]"+fmt.Sprintln(args...))
}
