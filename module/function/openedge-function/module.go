package main

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/function"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
)

// mo function module
type mo struct {
	cfg function.Config
	man *function.Manager
	rrs []*ruler
	log *logrus.Entry
}

// New creates a new module
func New(confDate string) (module.Module, error) {
	var cfg function.Config
	err := module.Load(&cfg, confDate)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defaults(&cfg)
	logger.Init(cfg.Logger, "module", cfg.Name)
	man, err := function.NewManager(cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	m := &mo{
		cfg: cfg,
		man: man,
		rrs: []*ruler{},
		log: logger.WithFields(),
	}
	for _, r := range cfg.Rules {
		f, err := man.Get(r.Compute.Function)
		if err != nil {
			m.Close()
			return nil, errors.Trace(err)
		}
		rr, err := create(r, cfg.Hub, f)
		if err != nil {
			m.Close()
			return nil, errors.Trace(err)
		}
		m.rrs = append(m.rrs, rr)
	}
	return m, nil
}

// Start starts module
func (m *mo) Start() error {
	m.log.Debug("module starting")
	for _, rr := range m.rrs {
		err := rr.start()
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// Close closes module
func (m *mo) Close() {
	defer m.log.Debug("module closed")
	for _, rr := range m.rrs {
		rr.close()
	}
	m.man.Close()
}

func defaults(cfg *function.Config) {
	if cfg.API.Address == "" {
		cfg.API.Address = module.GetEnv(module.EnvOpenEdgeMasterAPI)
	}
	if cfg.Hub.ClientID == "" {
		cfg.Hub.ClientID = cfg.Name
	}
	cfg.API.Username = cfg.Name
	cfg.API.Password = module.GetEnv(module.EnvOpenEdgeModuleToken)
}
