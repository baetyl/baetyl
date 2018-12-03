package function

import (
	"github.com/baidu/openedge/api"
	"github.com/baidu/openedge/logger"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
)

// Manager manages all functions
type Manager struct {
	cfg Config
	fcs map[string]*Function
	api *api.Client
	log *logrus.Entry
}

// NewManager loads all functions and return
func NewManager(c Config) (*Manager, error) {
	a, err := api.NewClient(c.API)
	if err != nil {
		return nil, errors.Trace(err)
	}
	m := &Manager{
		cfg: c,
		api: a,
		fcs: make(map[string]*Function),
		log: logger.WithFields("manager", "function"),
	}
	for _, fc := range c.Functions {
		m.fcs[fc.Name] = newFunction(m, fc)
	}
	return m, nil
}

// Get gets a function
func (m *Manager) Get(name string) (*Function, error) {
	f, ok := m.fcs[name]
	if !ok {
		return nil, errors.NotFoundf("Function (%s)", name)
	}
	return f, nil
}

// Close closes all function proc
func (m *Manager) Close() {
	m.log.Info("Function manager closing")
	defer m.log.Info("Function manager closed")
	for _, f := range m.fcs {
		f.close()
	}
}
