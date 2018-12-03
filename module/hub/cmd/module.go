package main

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/hub/broker"
	"github.com/baidu/openedge/module/hub/config"
	"github.com/baidu/openedge/module/hub/persist"
	"github.com/baidu/openedge/module/hub/rule"
	"github.com/baidu/openedge/module/hub/server"
	"github.com/baidu/openedge/module/hub/session"
	"github.com/juju/errors"
)

// mo openedge hub module
type mo struct {
	conf     *config.Config
	Rules    *rule.Manager
	Sessions *session.Manager
	broker   *broker.Broker
	servers  *server.Manager
	factory  *persist.Factory
}

// New creates a new module
func New(confFile string) (module.Module, error) {
	conf := new(config.Config)
	err := module.Load(conf, confFile)
	if err != nil {
		return nil, errors.Trace(err)
	}
	logger.Init(conf.Logger, "module", conf.Name)
	return &mo{
		conf: conf,
	}, nil
}

// Start starts module
func (m *mo) Start() (err error) {
	m.factory, err = persist.NewFactory(m.conf.Storage.Dir)
	if err != nil {
		return errors.Trace(err)
	}
	m.broker, err = broker.NewBroker(m.conf, m.factory)
	if err != nil {
		return errors.Trace(err)
	}
	m.Rules, err = rule.NewManager(m.conf.Subscriptions, m.broker)
	if err != nil {
		return errors.Trace(err)
	}
	m.Sessions, err = session.NewManager(m.conf, m.broker.Flow, m.Rules, m.factory)
	if err != nil {
		return errors.Trace(err)
	}
	m.servers, err = server.NewManager(m.conf.Listen, m.conf.Certificate, m.Sessions.Handle)
	if err != nil {
		return errors.Trace(err)
	}
	m.Rules.Start()
	m.servers.Start()
	return nil
}

// Close closes service
func (m *mo) Close() {
	if m.Rules != nil {
		m.Rules.Close()
	}
	if m.Sessions != nil {
		m.Sessions.Close()
	}
	if m.servers != nil {
		m.servers.Close()
	}
	if m.broker != nil {
		m.broker.Close()
	}
	if m.factory != nil {
		m.factory.Close()
	}
}
