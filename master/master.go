package master

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/api"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/baidubce/bce-sdk-go/util/log"
	cmap "github.com/orcaman/concurrent-map"
)

// Master master manages all modules and connects with cloud
type Master struct {
	cfg       Config
	appcfg    openedge.AppConfig
	server    *api.Server
	engine    engine.Engine
	services  cmap.ConcurrentMap
	accounts  cmap.ConcurrentMap
	infostats *infoStats
	log       logger.Logger
}

// New creates a new master
func New(pwd string, cfg Config, ver string) (*Master, error) {
	err := defaults(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to set default config: %s", err.Error())
	}
	err = os.MkdirAll(openedge.DefaultDBDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to make db directory: %s", err.Error())
	}
	log := logger.InitLogger(cfg.Logger, "openedge", "master")
	m := &Master{
		cfg:       cfg,
		log:       log,
		services:  cmap.New(),
		accounts:  cmap.New(),
		infostats: newInfoStats(pwd, cfg.Mode, ver, path.Join(openedge.DefaultDBDir, openedge.AppStatsFileName)),
	}
	log.Infof("mode: %s; grace: %d; pwd: %s", cfg.Mode, cfg.Grace, pwd)
	m.engine, err = engine.New(cfg.Mode, cfg.Grace, pwd, m.infostats)
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
	m, err = m.startServices()
	if err != nil && utils.FileExists(appBackupFile) {
		if err = m.rollback(); err != nil {
			m, err = m.startServices()
		}
	}
	return m, err
}

func (m *Master) startServices() (*Master, error) {
	_, _, _, err := m.prepareServices()
	if err != nil {
		m.Close()
		return nil, err
	}
	err = m.startAllServices(nil)
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
	m.stopAllServices(nil)
	if m.engine != nil {
		m.engine.Close()
	}
	return nil
}

func defaults(c *Config) error {
	if runtime.GOOS == "linux" {
		err := os.MkdirAll(path.Dir(openedge.DefaultSockFile), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to make directory of sock file: %s", err.Error())
		}
		c.Server.Address = "unix://" + openedge.DefaultSockFile
		utils.SetEnv(openedge.EnvMasterAPIKey, c.Server.Address)
	} else {
		if c.Server.Address == "" {
			c.Server.Address = "tcp://127.0.0.1:50050"
		}
		addr := c.Server.Address
		uri, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("failed to parse address of server: %s", err.Error())
		}
		if c.Mode == "docker" {
			parts := strings.SplitN(uri.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
		}
		utils.SetEnv(openedge.EnvMasterAPIKey, addr)
	}
	utils.SetEnv(openedge.EnvMasterAPIVersionKey, "v1")
	utils.SetEnv(openedge.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(openedge.EnvRunningModeKey, c.Mode)
	return nil
}
