package main

import (
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
)

// mo function module
type mo struct {
	cfg Config
	man *Manager
	rrs []*ruler
}

// New creates a new module
func New(confDate string) (module.Module, error) {
	var cfg Config
	err := module.Load(&cfg, confDate)
	if err != nil {
		return nil, err
	}
	defaults(&cfg)
	err = logger.Init(cfg.Logger, "module", cfg.UniqueName())
	if err != nil {
		return nil, err
	}
	man, err := NewManager(cfg)
	if err != nil {
		return nil, err
	}
	m := &mo{
		cfg: cfg,
		man: man,
		rrs: []*ruler{},
	}
	for _, r := range cfg.Rules {
		f, err := man.Get(r.Compute.Function)
		if err != nil {
			m.Close()
			return nil, err
		}
		rr, err := create(r, cfg.Hub, f)
		if err != nil {
			m.Close()
			return nil, err
		}
		m.rrs = append(m.rrs, rr)
	}
	return m, nil
}

// Start starts module
func (m *mo) Start() error {
	logger.Log.Debugf("module starting")

	for _, rr := range m.rrs {
		err := rr.start()
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes module
func (m *mo) Close() {
	defer logger.Log.Debugf("module closed")

	for _, rr := range m.rrs {
		rr.close()
	}
	m.man.Close()
}

func defaults(cfg *Config) {
	if cfg.API.Address == "" {
		cfg.API.Address = utils.GetEnv(module.EnvOpenEdgeMasterAPI)
	}
	cfg.API.Username = cfg.UniqueName()
	cfg.API.Password = utils.GetEnv(module.EnvOpenEdgeModuleToken)
}
