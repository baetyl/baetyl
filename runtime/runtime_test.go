package runtime_test

import (
	"strings"
	"testing"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/runtime"
	"github.com/stretchr/testify/assert"
	context "golang.org/x/net/context"
)

func TestRuntime(t *testing.T) {
	msg4k := &runtime.Message{Payload: []byte(strings.Repeat("a", 4*1024))}
	msg4m := &runtime.Message{Payload: []byte(strings.Repeat("a", 4*1024*1024))}
	msg8m := &runtime.Message{Payload: []byte(strings.Repeat("a", 8*1024*1024))}

	// server 4m by default
	sc := &runtime.Config{}
	err := module.Load(sc, `{"name":"test","server":{"address":"127.0.0.1:0"},"function":{"name":"test","handler":"dummy"}}`)
	assert.NoError(t, err)
	svr, err := runtime.NewServer(sc.Server, func(_ context.Context, m *runtime.Message) (*runtime.Message, error) {
		return m, nil
	})
	assert.NoError(t, err)

	// client 4m by default
	cc := runtime.NewClientConfig(svr.Address)
	cli, err := runtime.NewClient(cc)
	assert.NoError(t, err)

	out, err := cli.Handle(msg4k)
	assert.NoError(t, err)
	assert.Equal(t, msg4k.Payload, out.Payload)
	out, err = cli.Handle(msg4m)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (4194309 vs. 4194304)")
	assert.Nil(t, out)
	out, err = cli.Handle(msg8m)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (8388613 vs. 4194304)")
	assert.Nil(t, out)
	cli.Close()
	svr.Close()

	// server 10m
	sc.Server.Message.Length.Max = 10 * 1024 * 1024
	svr, err = runtime.NewServer(sc.Server, func(_ context.Context, m *runtime.Message) (*runtime.Message, error) {
		return m, nil
	})
	assert.NoError(t, err)

	// client 6m
	cc = runtime.NewClientConfig(svr.Address)
	cc.Message.Length.Max = 6 * 1024 * 1024
	cli, err = runtime.NewClient(cc)
	assert.NoError(t, err)

	out, err = cli.Handle(msg4k)
	assert.NoError(t, err)
	assert.Equal(t, msg4k.Payload, out.Payload)
	out, err = cli.Handle(msg4m)
	assert.NoError(t, err)
	assert.Equal(t, msg4m.Payload, out.Payload)
	out, err = cli.Handle(msg8m)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (8388613 vs. 6291456)")
	assert.Nil(t, out)
	cli.Close()

	// client 10m
	cc.Message.Length.Max = 10 * 1024 * 1024
	cli, err = runtime.NewClient(cc)
	assert.NoError(t, err)

	out, err = cli.Handle(msg4k)
	assert.NoError(t, err)
	assert.Equal(t, msg4k.Payload, out.Payload)
	out, err = cli.Handle(msg4m)
	assert.NoError(t, err)
	assert.Equal(t, msg4m.Payload, out.Payload)
	out, err = cli.Handle(msg8m)
	assert.NoError(t, err)
	assert.Equal(t, msg8m.Payload, out.Payload)
	cli.Close()
	svr.Close()
}
