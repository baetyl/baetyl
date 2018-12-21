// +build !windows

package master_test

import (
	"os"
	"testing"
	"time"

	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/module/api"
	"github.com/baidu/openedge/module/config"
	"github.com/stretchr/testify/assert"
)

func TestAPIUnix(t *testing.T) {
	os.MkdirAll("./var/", 0755)
	defer os.RemoveAll("./var/")
	addr := "unix://./var/test.sock"
	s, err := master.NewServer(&mockEngine{pass: true}, config.HTTPServer{Address: addr, Timeout: time.Minute})
	assert.NoError(t, err)
	defer s.Close()
	err = s.Start()
	assert.NoError(t, err)
	c, err := api.NewClient(config.HTTPClient{Address: addr, Timeout: time.Minute, KeepAlive: time.Minute})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	p, err := c.GetPortAvailable("127.0.0.1")
	assert.NoError(t, err)
	assert.NotZero(t, p)
	err = c.StartModule(&config.Module{Name: "name"})
	assert.NoError(t, err)
	err = c.StopModule(&config.Module{Name: "name"})
	assert.NoError(t, err)
}

func TestAPIUnixUnauthorized(t *testing.T) {
	os.MkdirAll("./var/", 0755)
	defer os.RemoveAll("./var/")
	addr := "unix://./var/test.sock"
	s, err := master.NewServer(&mockEngine{pass: false}, config.HTTPServer{Address: addr, Timeout: time.Minute})
	assert.NoError(t, err)
	defer s.Close()
	err = s.Start()
	assert.NoError(t, err)
	c, err := api.NewClient(config.HTTPClient{Address: addr, Timeout: time.Minute, KeepAlive: time.Minute, Username: "test"})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	_, err = c.GetPortAvailable("127.0.0.1")
	assert.EqualError(t, err, "[400] account (test) unauthorized")
	err = c.StartModule(&config.Module{Name: "name"})
	assert.EqualError(t, err, "[400] account (test) unauthorized")
	err = c.StopModule(&config.Module{Name: "name"})
	assert.EqualError(t, err, "[400] account (test) unauthorized")
}
