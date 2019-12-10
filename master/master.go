package master

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/api"
	"github.com/baetyl/baetyl/master/database"
	"github.com/baetyl/baetyl/master/engine"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	grpcapi "github.com/baetyl/baetyl/sdk/baetyl-go/api"
	cmap "github.com/orcaman/concurrent-map"
)

// Master master manages all modules and connects with cloud
type Master struct {
	cfg       Config
	ver       string
	pwd       string
	server    *api.Server
	engine    engine.Engine
	apiserver *grpcapi.Server
	services  cmap.ConcurrentMap
	database  database.DB
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

	err = os.MkdirAll(cfg.Database.Path, 0755)
	if err != nil {
		m.Close()
		return nil, fmt.Errorf("failed to make db directory: %s", err.Error())
	}
	m.database, err = database.New(database.Conf{Driver: cfg.Database.Driver, Source: path.Join(cfg.Database.Path, "kv.db")})
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("db inited")

	m.apiserver, err = grpcapi.NewServer(cfg.API, m)
	if err != nil {
		m.Close()
		return nil, err
	}
	m.apiserver.RegisterKVService(api.NewKVService(m.database))
	err = m.apiserver.Start()
	if err != nil {
		m.Close()
		return nil, err
	}
	log.Infoln("api server started")

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
	if m.apiserver != nil {
		m.apiserver.Close()
		m.log.Infoln("api server stopped")
	}
	if m.database != nil {
		m.database.Close()
		m.log.Infoln("db closed")
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
