package master

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
)

// Version of master
var Version string

// Master master manages all modules and connects with cloud
type Master struct {
	inicfg   Config
	precfg   *DynamicConfig
	curcfg   *DynamicConfig
	server   *Server
	engine   engine.Engine
	services cmap.ConcurrentMap
	accounts cmap.ConcurrentMap
	pwd      string
	log      logger.Logger
}

// New creates a new master
func New(pwd, confpath string) (*Master, error) {
	var cfg Config
	err := utils.LoadYAML(path.Join(pwd, confpath), &cfg)
	if err != nil {
		return nil, err
	}
	err = defaults(&cfg)
	if err != nil {
		return nil, err
	}
	log, err := logger.InitLogger(&cfg.Logger, "openedge", "master")
	if err != nil {
		return nil, err
	}
	m := &Master{
		services: cmap.New(),
		accounts: cmap.New(),
		inicfg:   cfg,
		pwd:      pwd,
		log:      log,
	}
	log.Infof("mode: %s; grace: %d; pwd: %s", cfg.Mode, cfg.Grace, m.pwd)
	m.engine, err = engine.New(cfg.Mode, cfg.Grace, m.pwd)
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("engine started")
	err = m.initServer()
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("server started")
	err = m.initServices()
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
	utils.SetEnv(openedge.EnvServiceModeKey, c.Mode)
	return nil
}
