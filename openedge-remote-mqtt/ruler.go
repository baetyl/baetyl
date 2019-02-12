package main

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
)

type ruler struct {
	r      *Rule
	hub    *mqtt.Dispatcher
	remote *mqtt.Dispatcher
	log    logger.Logger
}

func create(r Rule, hub, remote mqtt.ClientInfo) *ruler {
	if remote.ClientID == "" {
		remote.ClientID = r.Remote.Name
	}
	hub.ClientID = remote.ClientID
	hub.Subscriptions = r.Hub.Subscriptions
	remote.Subscriptions = r.Remote.Subscriptions
	return &ruler{
		r:      &r,
		hub:    mqtt.NewDispatcher(hub),
		remote: mqtt.NewDispatcher(remote),
		log:    logger.WithField("rule", remote.ClientID),
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
