package main

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
)

type ruler struct {
	rule   *Rule
	hub    *mqtt.Dispatcher
	remote *mqtt.Dispatcher
	log    logger.Logger
}

func create(rule Rule, hub, remote mqtt.ClientInfo) *ruler {
	defaults(&rule, &hub, &remote)
	log := logger.WithField("rule", rule.Remote.Name)
	return &ruler{
		rule:   &rule,
		hub:    mqtt.NewDispatcher(hub, log),
		remote: mqtt.NewDispatcher(remote, log),
		log:    log,
	}
}

func (rr *ruler) start() error {
	hubHandler := mqtt.NewHandlerWrapper(
		func(p *packet.Publish) error {
			return rr.remote.Send(p)
		},
		func(p *packet.Puback) error {
			return rr.remote.Send(p)
		},
		func(e error) {
			rr.log.Errorln("hub error:", e.Error())
		},
	)
	if err := rr.hub.Start(hubHandler); err != nil {
		return err
	}
	remoteHandler := mqtt.NewHandlerWrapper(
		func(p *packet.Publish) error {
			return rr.hub.Send(p)
		},
		func(p *packet.Puback) error {
			return rr.hub.Send(p)
		},
		func(e error) {
			rr.log.Errorln("remote error:", e.Error())
		},
	)
	if err := rr.remote.Start(remoteHandler); err != nil {
		return err
	}
	return nil
}

func (rr *ruler) close() {
	rr.hub.Close()
	rr.remote.Close()
}

func defaults(rule *Rule, hub, remote *mqtt.ClientInfo) {
	// set remote client id
	// rules[].remote.clientid > remotes[].clientid > rules[].remote.name
	if rule.Remote.ClientID != "" {
		remote.ClientID = rule.Remote.ClientID
	} else if remote.ClientID == "" {
		remote.ClientID = rule.Remote.Name
	}
	// set hub client id
	// rules[].hub.clientid > remote.ClientID
	if rule.Hub.ClientID != "" {
		hub.ClientID = rule.Hub.ClientID
	} else {
		hub.ClientID = remote.ClientID
	}
	hub.Subscriptions = rule.Hub.Subscriptions
	remote.Subscriptions = rule.Remote.Subscriptions
}
