package master

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/api"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
)

// Version of master
var Version string

// Master master manages all modules and connects with cloud
type Master struct {
	cfg      Config
	appcfg   openedge.AppConfig
	server   *api.Server
	engine   engine.Engine
	services cmap.ConcurrentMap
	accounts cmap.ConcurrentMap
	context  cmap.ConcurrentMap
	pwd      string
	log      logger.Logger
}

// New creates a new master
func New(pwd string, cfg *Config) (*Master, error) {
	log, err := logger.InitLogger(&cfg.Logger, "openedge", "master")
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %s", err.Error())
	}
	err = defaults(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to set default config: %s", err.Error())
	}
	m := &Master{
		cfg:      *cfg,
		pwd:      pwd,
		log:      log,
		services: cmap.New(),
		accounts: cmap.New(),
		context:  cmap.New(),
	}
	log.Infof("mode: %s; grace: %d; pwd: %s", cfg.Mode, cfg.Grace, m.pwd)
	m.engine, err = engine.New(cfg.Mode, cfg.Grace, m.pwd)
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("engine started")
	m.server, err = api.New(m.cfg.Server, m)
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("server started")
	err = m.prepareServices()
	if err != nil {
		m.Close()
		return nil, err
	}
	err = m.startAllServices()
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("services started")
	return m, nil
}

// Close closes agent
func (m *Master) Close() error {
	defer m.log.Infoln("master stopped")
	if m.server != nil {
		m.server.Close()
	}
	m.stopAllServices()
	if m.engine != nil {
		m.engine.Close()
	}
	return nil
}

func defaults(c *Config) error {
	if runtime.GOOS == "linux" {
		err := os.MkdirAll("/var/run", os.ModePerm)
		if err != nil {
			logger.WithError(err).Errorf("failed to make dir: /var/run")
		}
		c.Server.Address = "unix:///var/run/openedge.sock"
		utils.SetEnv(openedge.EnvMasterAPIKey, c.Server.Address)
	} else {
		if c.Server.Address == "" {
			c.Server.Address = "tcp://127.0.0.1:50050"
		}
		addr := c.Server.Address
		uri, err := url.Parse(addr)
		if err != nil {
			return err
		}
		if c.Mode == "docker" {
			parts := strings.SplitN(uri.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
		}
		utils.SetEnv(openedge.EnvMasterAPIKey, addr)
	}
	utils.SetEnv(openedge.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(openedge.EnvRunningModeKey, c.Mode)
	return nil
}
