package main

import (
	"fmt"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/logger"
)

// mo bridge module of mqtt servers
type mo struct {
	cfg Config
	rrs []*ruler
	log *logger.Entry
}

// New create a new module
func New(confFile string) (module.Module, error) {
	var cfg Config
	err := module.Load(&cfg, confFile)
	if err != nil {
		return nil, err
	}
	logger.Init(cfg.Logger, "module", cfg.Name)
	remotes := make(map[string]Remote)
	for _, remote := range cfg.Remotes {
		remotes[remote.Name] = remote
	}
	rulers := make([]*ruler, 0)
	for _, rule := range cfg.Rules {
		remote, ok := remotes[rule.Remote.Name]
		if !ok {
			return nil, fmt.Errorf("remote (%s) not found", rule.Remote.Name)
		}
		rulers = append(rulers, create(rule, cfg.Hub, remote.MQTTClient))
	}
	return &mo{
		cfg: cfg,
		rrs: rulers,
		log: logger.WithFields(),
	}, nil
}

// Start starts module
func (m *mo) Start() error {
	for _, ruler := range m.rrs {
		err := ruler.start()
		if err != nil {
			m.log.WithError(err).Errorf("failed to start rule")
			return err
		}
	}
	return nil
}

// Close closes module
func (m *mo) Close() {
	for _, ruler := range m.rrs {
		ruler.close()
	}
}
