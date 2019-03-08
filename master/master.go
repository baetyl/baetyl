package master

import (
	"fmt"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/api"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
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
