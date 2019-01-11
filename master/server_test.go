package master

import (
	"fmt"
	"testing"
	"time"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/master"
	"github.com/stretchr/testify/assert"
)

type mockAPI struct {
	pass bool
}

func (a *mockAPI) stats() *master.Stats {
	fmt.Println("stats")
	return master.NewStats()
}

func (a *mockAPI) reload(file string) error {
	fmt.Println("reload", file)
	return nil
}

func (a *mockAPI) authModule(username, password string) bool {
	fmt.Println("authModule")
	return a.pass
}

func (a *mockAPI) startModule(_ config.Module) error {
	fmt.Println("startModule")
	return nil
}

func (a *mockAPI) stopModule(_ string) error {
	fmt.Println("stopModule")
	return nil
}

func TestAPIHttp(t *testing.T) {
	s, err := NewServer(&mockAPI{pass: true}, config.HTTPServer{Address: "tcp://127.0.0.1:0", Timeout: time.Minute})
	assert.NoError(t, err)
	defer s.Close()
	err = s.Start()
	assert.NoError(t, err)
	c, err := master.NewClient(config.HTTPClient{Address: "tcp://" + s.Addr, Timeout: time.Minute, KeepAlive: time.Minute})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	p, err := c.GetPortAvailable("127.0.0.1")
	assert.NoError(t, err)
	assert.NotZero(t, p)
	err = c.StartModule(&config.Module{Name: "name"})
	assert.NoError(t, err)
	err = c.StopModule(&config.Module{Name: "name"})
	assert.NoError(t, err)
	stats, err := c.Stats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	err = c.Reload("var/db/v1.zip")
	assert.NoError(t, err)
}

func TestAPIHttpUnauthorized(t *testing.T) {
	s, err := NewServer(&mockAPI{pass: false}, config.HTTPServer{Address: "tcp://127.0.0.1:0", Timeout: time.Minute})
	assert.NoError(t, err)
	defer s.Close()
	err = s.Start()
	assert.NoError(t, err)
	c, err := master.NewClient(config.HTTPClient{Address: "tcp://" + s.Addr, Timeout: time.Minute, KeepAlive: time.Minute, Username: "test"})
	assert.NoError(t, err)
	assert.NotNil(t, c)
	_, err = c.GetPortAvailable("127.0.0.1")
	assert.EqualError(t, err, "[400] account (test) unauthorized")
	err = c.StartModule(&config.Module{Name: "name"})
	assert.EqualError(t, err, "[400] account (test) unauthorized")
	err = c.StopModule(&config.Module{Name: "name"})
	assert.EqualError(t, err, "[400] account (test) unauthorized")
	stats, err := c.Stats()
	assert.EqualError(t, err, "[400] account (test) unauthorized")
	assert.Nil(t, stats)
	err = c.Reload("var/db/v1.zip")
	assert.EqualError(t, err, "[400] account (test) unauthorized")
}
