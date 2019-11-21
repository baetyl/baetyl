package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
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

	log := newMockLogger()
	m := mockMaster{database: db, l: log}

	conf := Conf{Address: "baetyl"}
	apiServer, err := NewAPIServer(conf, &m)
	assert.Error(t, err)
	if apiServer != nil {
		apiServer.Close()
	}

	conf = Conf{Address: "tcp://127.0.0.1:10000000"}
	apiServer, err = NewAPIServer(conf, &m)
	assert.Error(t, err)
	if apiServer != nil {
		apiServer.Close()
	}

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
		apiServer, err = NewAPIServer(Conf{conf.server}, &m)
		assert.NoError(t, err)
		assert.Contains(t, log.GetRecords(), fmt.Sprintf("[Infof]api server is listening at: %s", conf.server))

		conn, err := grpc.Dial(conf.client, grpc.WithInsecure())
		assert.NoError(t, err)
		client := baetyl.NewKVServiceClient(conn)
		assert.NotEmpty(t, client)

		ctx := context.Background()
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
	}
}

type mockMaster struct {
	database database.DB
	l        *mockLogger
}

func (m *mockMaster) Auth(u, p string) bool {
	return false
}

func (m *mockMaster) InspectSystem() ([]byte, error) {
	return nil, nil
}

func (m *mockMaster) UpdateSystem(trace, tp, target string) error {
	return nil
}

func (m *mockMaster) ReportInstance(serviceName, instanceName string, partialStats engine.PartialStats) error {
	return nil
}

func (m *mockMaster) StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	return nil
}
func (m *mockMaster) StopInstance(serviceName, instanceName string) error {
	return nil
}

func (m *mockMaster) Logger() logger.Logger {
	return m.l
}

func (m *mockMaster) SetKV(kv *baetyl.KV) error {
	return m.database.Set(kv)
}
func (m *mockMaster) GetKV(key []byte) (*baetyl.KV, error) {
	return m.database.Get(key)
}
func (m *mockMaster) DelKV(key []byte) error {
	return m.database.Del(key)
}
func (m *mockMaster) ListKV(prefix []byte) (*baetyl.KVs, error) {
	return m.database.List(prefix)
}

type mockLogger struct {
	records []string
	data    map[string]interface{}
	err     error
	sync.RWMutex
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		data:    map[string]interface{}{},
		records: []string{},
	}
}

func (l *mockLogger) WithField(key string, value interface{}) logger.Logger {
	l.Lock()
	defer l.Unlock()
	l.data[key] = value
	return l
}
func (l *mockLogger) WithError(err error) logger.Logger {
	l.Lock()
	defer l.Unlock()
	l.err = err
	return l
}
func (l *mockLogger) Debugf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Debugf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Infof(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Infof]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Warnf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Warnf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Errorf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Errorf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Fatalf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Fatalf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Debugln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Debugln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Infoln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Infoln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Warnln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Warnln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Errorln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Errorln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Fatalln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Fatalln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) GetRecords() []string {
	l.Lock()
	defer l.Unlock()
	return l.records
}
