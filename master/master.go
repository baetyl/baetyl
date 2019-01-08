package master

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/baidu/openedge/engine"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
)

// Master master manages all modules and connects with cloud
type Master struct {
	conf    Config
	context engine.Context
	engine  *engine.Engine
	server  *Server
}

// New creates a new master
func New(confDate string) (*Master, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	m := &Master{}
	err = module.Load(&m.conf, confDate)
	if err != nil {
		return nil, err
	}
	err = defaults(&m.conf)
	if err != nil {
		return nil, err
	}
	logger.Init(m.conf.Logger, "openedge", "master")
	logger.Log.Debugln("work dir:", pwd)
	m.context = engine.Context{
		PWD:   pwd,
		Mode:  m.conf.Mode,
		Grace: m.conf.Grace,
	}
	m.engine, err = engine.New(&m.context)
	if err != nil {
		return nil, err
	}
	m.server, err = NewServer(m, m.conf.API)
	if err != nil {
		return nil, err
	}
	err = m.server.Start()
	if err != nil {
		m.Close()
		return nil, err
	}
	return m, nil
}

// Start starts agent
func (m *Master) Start() error {
	if err := m.loadConfig(); err != nil {
		return err
	}
	return m.engine.StartAll(m.conf.Modules)
}

// Close closes agent
func (m *Master) Close() {
	if m.engine != nil {
		m.engine.StopAll()
	}
	if err := m.server.Close(); err != nil {
		logger.WithError(err).Errorf("failed to close api server")
	}
}

func (m *Master) authModule(username, password string) bool {
	return m.engine.Authenticate(username, password)
}

func (m *Master) startModule(module config.Module) error {
	return m.engine.Start(module)
}

func (m *Master) stopModule(module string) error {
	return m.engine.Stop(module)
}

func defaults(c *Config) error {
	if runtime.GOOS == "linux" {
		err := os.MkdirAll("var/run", os.ModePerm)
		if err != nil {
			logger.WithError(err).Errorf("failed to make dir: var/run")
		}
		c.API.Address = "unix://var/run/openedge.sock"
		utils.SetEnv(module.EnvOpenEdgeMasterAPI, c.API.Address)
	} else {
		if c.API.Address == "" {
			c.API.Address = "tcp://127.0.0.1:50050"
		}
		addr := c.API.Address
		uri, err := url.Parse(addr)
		if err != nil {
			return err
		}
		if c.Mode == "docker" {
			parts := strings.SplitN(uri.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
		}
		utils.SetEnv(module.EnvOpenEdgeMasterAPI, addr)
	}
	utils.SetEnv(module.EnvOpenEdgeHostOS, runtime.GOOS)
	utils.SetEnv(module.EnvOpenEdgeModuleMode, c.Mode)
	return nil
}
