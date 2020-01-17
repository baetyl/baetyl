package master

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/api"
	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	cmap "github.com/orcaman/concurrent-map"
)

// Master master manages all modules and connects with cloud
type Master struct {
	cfg       Config
	ver       string
	pwd       string
	server    *api.Server
	engine    engine.Engine
	services  cmap.ConcurrentMap
	accounts  cmap.ConcurrentMap
	infostats *infoStats
	sig       chan os.Signal
	log       logger.Logger
}

// New creates a new master
func New(pwd string, cfg Config, ver string, revision string) (*Master, error) {
	err := os.MkdirAll(baetyl.DefaultDBDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to make db directory: %s", err.Error())
	}

	log := logger.InitLogger(cfg.Logger, "baetyl", "master")
	m := &Master{
		cfg:       cfg,
		ver:       ver,
		pwd:       pwd,
		log:       log,
		sig:       make(chan os.Signal, 1),
		services:  cmap.New(),
		accounts:  cmap.New(),
		infostats: newInfoStats(pwd, cfg.Mode, ver, revision, path.Join(baetyl.DefaultDBDir, baetyl.AppStatsFileName)),
	}
	log.Infof("mode: %s; grace: %d; pwd: %s; api: %s", cfg.Mode, cfg.Grace, pwd, cfg.Server.Address)

	opts := engine.Options{
		Grace:      cfg.Grace,
		Pwd:        pwd,
		APIVersion: cfg.Docker.APIVersion,
	}
	m.engine, err = engine.New(cfg.Mode, m.infostats, opts)
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

	// TODO: implement recover logic when master restarts
	// Now it will stop all old services
	m.engine.Recover()

	// start application
	err = m.UpdateAPP("", "")
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
	select {
	case m.sig <- syscall.SIGQUIT:
	default:
	}
	return nil
}

// Wait waits until master closes
func (m *Master) Wait() error {
	signal.Notify(m.sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-m.sig
	return nil
}
