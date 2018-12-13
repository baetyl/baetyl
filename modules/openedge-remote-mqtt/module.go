package main

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/sirupsen/logrus"
)

// mo bridge module of mqtt servers
type mo struct {
	conf   Config
	hub    *mqtt.Dispatcher
	remote *mqtt.Dispatcher
	log    *logrus.Entry
}

// New create a new module
func New(confFile string) (module.Module, error) {
	var conf Config
	err := module.Load(&conf, confFile)
	if err != nil {
		return nil, err
	}
	logger.Init(conf.Logger, "module", conf.Name)
	hub := mqtt.NewDispatcher(conf.Hub)
	remote := mqtt.NewDispatcher(conf.Remote)
	return &mo{
		conf:   conf,
		hub:    hub,
		remote: remote,
		log:    logger.WithFields(),
	}, nil
}

// Start starts module
func (m *mo) Start() error {
	err := m.hub.Start(func(pkt packet.Generic) {
		err := m.remote.Send(pkt)
		if err != nil {
			m.log.WithError(err).Errorf("failed to send packet to remote")
		}
	})
	if err != nil {
		m.log.WithError(err).Errorf("failed to start hub dispatcher")
		return err
	}
	err = m.remote.Start(func(pkt packet.Generic) {
		err := m.hub.Send(pkt)
		if err != nil {
			m.log.WithError(err).Errorf("failed to send packet to hub")
		}
	})
	if err != nil {
		m.log.WithError(err).Errorf("failed to start remote dispatcher")
		return err
	}
	return nil
}

// Close closes module
func (m *mo) Close() {
	if m.hub != nil {
		err := m.hub.Close()
		if err != nil {
			m.log.WithError(err).Errorf("failed to close hub dispatcher")
		}
	}
	if m.remote != nil {
		err := m.remote.Close()
		if err != nil {
			m.log.WithError(err).Errorf("failed to close remote dispatcher")
		}
	}
}
