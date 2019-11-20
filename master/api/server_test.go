package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/master/engine"
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

	log := newMockLogger()
	conf := Conf{Address: "baetyl"}
	m := mockMaster{DB: db, l: log}

	_, err = NewAPIServer(conf, &m)
	assert.Error(t, err)

	conf = Conf{Address: "tcp://127.0.0.1:10000000"}
	apiServer, err := NewAPIServer(conf, &m)
	assert.Error(t, err)

	conf = Conf{Address: "tcp://127.0.0.1:50060"}
	apiServer, err = NewAPIServer(conf, &m)
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

type mockMaster struct {
	database.DB
	l *mockLogger
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

type mockLogger struct {
	records []string
	data    map[string]interface{}
	err     error
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		data:    map[string]interface{}{},
		records: []string{},
	}
}

func (l *mockLogger) WithField(key string, value interface{}) logger.Logger {
	l.data[key] = value
	return l
}
func (l *mockLogger) WithError(err error) logger.Logger {
	l.err = err
	return l
}
func (l *mockLogger) Debugf(format string, args ...interface{}) {
	l.records = append(l.records, "[Debugf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Infof(format string, args ...interface{}) {
	l.records = append(l.records, "[Infof]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Warnf(format string, args ...interface{}) {
	l.records = append(l.records, "[Warnf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Errorf(format string, args ...interface{}) {
	l.records = append(l.records, "[Errorf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Fatalf(format string, args ...interface{}) {
	l.records = append(l.records, "[Fatalf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Debugln(args ...interface{}) {
	l.records = append(l.records, "[Debugln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Infoln(args ...interface{}) {
	l.records = append(l.records, "[Infoln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Warnln(args ...interface{}) {
	l.records = append(l.records, "[Warnln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Errorln(args ...interface{}) {
	l.records = append(l.records, "[Errorln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Fatalln(args ...interface{}) {
	l.records = append(l.records, "[Fatalln]"+fmt.Sprintln(args...))
}
