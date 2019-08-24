package baetyl

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestFunctionClient(t *testing.T) {
	t.Skip("local test")

	cc := FunctionClientConfig{
		Address: "127.0.0.1:50051",
	}
	err := utils.UnmarshalJSON([]byte("{\"address\":\"127.0.0.1:50051\"}"), &cc)
	assert.NoError(t, err)
	cli, err := NewFClient(cc)
	assert.NoError(t, err)

	fn := "sayhi3"
	msg := &FunctionMessage{
		// Payload: []byte("{\"bytes\":\"a\"}"),
		// Payload:          []byte("{\"err\":\"error\"}"),
		Payload:          []byte("a"),
		FunctionName:     fn,
		FunctionInvokeID: "invokeid",
	}
	start := time.Now()
	out, err := cli.Call(msg)
	fmt.Printf("%s elapsed time: %v\n", t.Name(), time.Since(start))
	assert.NoError(t, err)
	res := make(map[string]interface{})
	err = json.Unmarshal(out.Payload, &res)
	assert.NoError(t, err)
	assert.Equal(t, "Hello Baetyl", res["Say"])
	assert.Equal(t, "a", res["bytes"])
	assert.Equal(t, fn, res["functionName"])
	assert.Equal(t, "invokeid", res["functionInvokeID"])
}

func TestFunctionCall(t *testing.T) {
	msg := &FunctionMessage{FunctionName: "happy", QOS: 1, Topic: "t", Payload: []byte(`{"v":"a"}`), FunctionInvokeID: "x"}
	msg4k := &FunctionMessage{FunctionName: "test", Payload: []byte(strings.Repeat("a", 4*1024))}
	msg4m := &FunctionMessage{FunctionName: "test", Payload: []byte(strings.Repeat("a", 4*1024*1024))}
	msg8m := &FunctionMessage{FunctionName: "test", Payload: []byte(strings.Repeat("a", 8*1024*1024))}
	call := func(c context.Context, m *FunctionMessage) (*FunctionMessage, error) {
		if m.FunctionName == "happy" {
			assert.NotNil(t, m.Payload)
			out := make(map[string]string)
			err := json.Unmarshal(m.Payload, &out)
			assert.NoError(t, err)
			out["qos"] = fmt.Sprintf("%d", m.QOS)
			out["topic"] = m.Topic
			out["fn"] = m.FunctionName
			out["fii"] = m.FunctionInvokeID
			m.Payload, err = json.Marshal(out)
			assert.NoError(t, err)
		} else {
			assert.Equal(t, "test", m.FunctionName)
		}
		return m, nil
	}

	// server 4m by default
	sc := FunctionServerConfig{}
	err := utils.UnmarshalJSON([]byte(`{"address":"127.0.0.1:0"}`), &sc)
	assert.NoError(t, err)
	svr, err := NewFServer(sc, call)
	assert.NoError(t, err)

	// client 4m by default
	cc := FunctionClientConfig{}
	err = utils.UnmarshalJSON([]byte("{\"address\":\""+svr.addr+"\"}"), &cc)
	assert.NoError(t, err)
	cli, err := NewFClient(cc)
	assert.NoError(t, err)

	out, err := cli.Call(msg)
	assert.NoError(t, err)
	res := make(map[string]string)
	err = json.Unmarshal(out.Payload, &res)
	assert.NoError(t, err)
	assert.Equal(t, "1", res["qos"])
	assert.Equal(t, "t", res["topic"])
	assert.Equal(t, "happy", res["fn"])
	assert.Equal(t, "x", res["fii"])
	assert.Equal(t, "a", res["v"])

	out, err = cli.Call(msg4k)
	assert.NoError(t, err)
	assert.Equal(t, msg4k.Payload, out.Payload)
	out, err = cli.Call(msg4m)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (4194315 vs. 4194304)")
	assert.Nil(t, out)
	out, err = cli.Call(msg8m)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (8388619 vs. 4194304)")
	assert.Nil(t, out)
	cli.Close()
	svr.Close()

	// server 10m
	sc.Message.Length.Max = 10 * 1024 * 1024
	svr, err = NewFServer(sc, call)
	assert.NoError(t, err)

	// client 6m
	cc = FunctionClientConfig{Address: svr.addr, Timeout: time.Second}
	cc.Message.Length.Max = 6 * 1024 * 1024
	cli, err = NewFClient(cc)
	assert.NoError(t, err)

	out, err = cli.Call(msg4k)
	assert.NoError(t, err)
	assert.Equal(t, msg4k.Payload, out.Payload)
	out, err = cli.Call(msg4m)
	assert.NoError(t, err)
	assert.Equal(t, msg4m.Payload, out.Payload)
	out, err = cli.Call(msg8m)
	assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = grpc: received message larger than max (8388619 vs. 6291456)")
	assert.Nil(t, out)
	cli.Close()

	// client 10m
	cc.Message.Length.Max = 10 * 1024 * 1024
	cli, err = NewFClient(cc)
	assert.NoError(t, err)

	out, err = cli.Call(msg4k)
	assert.NoError(t, err)
	assert.Equal(t, msg4k.Payload, out.Payload)
	out, err = cli.Call(msg4m)
	assert.NoError(t, err)
	assert.Equal(t, msg4m.Payload, out.Payload)
	out, err = cli.Call(msg8m)
	assert.NoError(t, err)
	assert.Equal(t, msg8m.Payload, out.Payload)
	cli.Close()
	svr.Close()
}
