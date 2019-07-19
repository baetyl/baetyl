package master

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/api"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/protocol/http"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
)

// Master master manages all modules and connects with cloud
type Master struct {
	cfg       Config
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
	log.Infof("mode: %s; grace: %d; pwd: %s; api: %s", cfg.Mode, cfg.Grace, pwd, cfg.Server.Address)
	m.engine, err = engine.New(cfg.Mode, cfg.Grace, pwd, m.infostats)
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("engine started")
	sc := http.ServerInfo{
		Address:     m.cfg.Server.Address,
		Timeout:     m.cfg.Server.Timeout,
		Certificate: m.cfg.Server.Certificate,
	}
	m.server, err = api.New(sc, m)
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("server started")
	// TODO: implement recover logic when master restarts
	// Now it will stop all old services
	m.engine.Recover()
	// start application
	err = m.UpdateSystem("")
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("services started")
	return m, nil
}

// Close closes agent
func (m *Master) Close() error {
	if m.server != nil {
		m.server.Close()
		m.log.Infoln("server stopped")
	}
	m.stopServices(map[string]struct{}{})
	if m.engine != nil {
		m.engine.Close()
		m.log.Infoln("engine stopped")
	}
	return nil
}

func defaults(c *Config) error {
	addr := c.Server.Address
	url, err := utils.ParseURL(addr)
	if err != nil {
		return fmt.Errorf("failed to parse address of server: %s", err.Error())
	}

	if runtime.GOOS != "linux" && url.Scheme == "unix" {
		return fmt.Errorf("unix domain socket only support on linux, please to use tcp socket")
	}
	if url.Scheme != "unix" && url.Scheme != "tcp" {
		return fmt.Errorf("only support unix domian socket or tcp socket")
	}

	// address in container
	if url.Scheme == "unix" {
		sock, err := filepath.Abs(url.Host)
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Dir(sock), 0755)
		if err != nil {
			return err
		}
		utils.SetEnv(openedge.EnvMasterHostSocketKey, sock)
		if c.Mode == "native" {
			utils.SetEnv(openedge.EnvMasterAPIKey, "unix://"+openedge.DefaultSockFile)
		} else {
			utils.SetEnv(openedge.EnvMasterAPIKey, "unix:///"+openedge.DefaultSockFile)
		}
	} else {
		if c.Mode == "native" {
			utils.SetEnv(openedge.EnvMasterAPIKey, addr)
		} else {
			parts := strings.SplitN(url.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
			utils.SetEnv(openedge.EnvMasterAPIKey, addr)
		}
	}
	utils.SetEnv(openedge.EnvMasterAPIVersionKey, "v1")
	utils.SetEnv(openedge.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(openedge.EnvRunningModeKey, c.Mode)
	return nil
}
