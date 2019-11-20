package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"google.golang.org/grpc"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func Test_KVServer(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := database.New(database.Conf{Driver: "sqlite3", Source: path.Join(dir, "kv.db")})
	assert.NoError(t, err)
	defer db.Close()

	log := newMockLogger()
	err = NewKVServer(log, db)
	assert.NoError(t, err)
	assert.Equal(t, []string{"[Infoln]kv server is listening at: tcp://127.0.0.1:50060\n"}, log.records)

	conn, err := grpc.Dial("127.0.0.1:50060", grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()
	client := baetyl.NewKVClient(conn)
	assert.NotEmpty(t, client)

	ctx := context.Background()
	// GetKV empty
	_, err = client.GetKV(ctx, &baetyl.KVMessage{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = key is empty")

	// GetKV empty
	resp, err := client.GetKV(ctx, &baetyl.KVMessage{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = key is empty")

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = key is empty")

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("aa")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = value is empty")

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("aa"), Value: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = value is empty")

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("aa"), Value: []byte("aadata")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// GetKV
	resp, err = client.GetKV(ctx, &baetyl.KVMessage{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Equal(t, resp.Key, []byte("aa"))
	assert.Equal(t, resp.Value, []byte("aadata"))

	// DelKV
	resp, err = client.DelKV(ctx, &baetyl.KVMessage{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// DelKV
	resp, err = client.DelKV(ctx, &baetyl.KVMessage{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = key is empty")

	// GetKV empty
	resp, err = client.GetKV(ctx, &baetyl.KVMessage{Key: []byte("aa")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("/root/a"), Value: []byte("/root/ax")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("/root/b"), Value: []byte("/root/bx")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// PutKV
	resp, err = client.PutKV(ctx, &baetyl.KVMessage{Key: []byte("/rootx/a"), Value: []byte("/rootx/ax")})
	assert.NoError(t, err)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Value)

	// ListKV
	respa, err := client.ListKV(ctx, &baetyl.KVMessage{Key: []byte("/root")})
	assert.NoError(t, err)
	assert.Len(t, respa.Kvs, 2)
	assert.Equal(t, respa.Kvs[0].Key, []byte("/root/a"))
	assert.Equal(t, respa.Kvs[1].Key, []byte("/root/b"))
	assert.Equal(t, respa.Kvs[0].Value, []byte("/root/ax"))
	assert.Equal(t, respa.Kvs[1].Value, []byte("/root/bx"))

	// ListKV
	respa, err = client.ListKV(ctx, &baetyl.KVMessage{Key: []byte("/rootx")})
	assert.NoError(t, err)
	assert.Len(t, respa.Kvs, 1)
	assert.Equal(t, respa.Kvs[0].Key, []byte("/rootx/a"))
	assert.Equal(t, respa.Kvs[0].Value, []byte("/rootx/ax"))

	// ListKV
	respa, err = client.ListKV(ctx, &baetyl.KVMessage{Key: []byte("")})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = Unknown desc = key is empty")
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
