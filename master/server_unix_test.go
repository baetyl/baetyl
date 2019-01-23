// +build !windows

package master

import (
	"os"
	"testing"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/protocol/jrpc"
	"github.com/stretchr/testify/assert"
)

func TestAPIUnix(t *testing.T) {
	os.MkdirAll("./var/", 0755)
	defer os.RemoveAll("./var/")
	addr := "unix://./var/test.sock"
	s, err := newServer(addr, &mockMaster{pass: true})
	assert.NoError(t, err)
	defer s.s.Close()
	c, err := jrpc.NewClient(addr)
	assert.NoError(t, err)
	defer c.Close()
	err = c.Call(openedge.CallUpdateSystem, &openedge.UpdateSystemRequest{}, &openedge.UpdateSystemResponse{})
	assert.NoError(t, err)

	// p, err := c.GetPortAvailable("127.0.0.1")
	// assert.NoError(t, err)
	// assert.NotZero(t, p)
	// err = c.StartModule(&config.Module{Name: "name"})
	// assert.NoError(t, err)
	// err = c.StopModule(&config.Module{Name: "name"})
	// assert.NoError(t, err)
}

// func TestAPIUnixUnauthorized(t *testing.T) {
// 	s, err := newServer(&mockMaster{pass: false}, config.HTTPServer{Address: "tcp://127.0.0.1:0", Timeout: time.Minute})
// 	assert.NoError(t, err)
// 	defer s.Close()
// 	err = s.Start()
// 	assert.NoError(t, err)
// 	c, err := master.NewClient(config.HTTPClient{Address: "tcp://" + s.Addr, Timeout: time.Minute, KeepAlive: time.Minute, Username: "test"})
// 	assert.NoError(t, err)
// 	assert.NotNil(t, c)
// 	_, err = c.GetPortAvailable("127.0.0.1")
// 	assert.EqualError(t, err, "[400] account (test) unauthorized")
// 	err = c.StartModule(&config.Module{Name: "name"})
// 	assert.EqualError(t, err, "[400] account (test) unauthorized")
// 	err = c.StopModule(&config.Module{Name: "name"})
// 	assert.EqualError(t, err, "[400] account (test) unauthorized")
// }
