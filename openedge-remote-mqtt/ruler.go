package main

import (
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/mqtt"
	"github.com/docker/distribution/uuid"
)

type ruler struct {
	r      *Rule
	hub    *mqtt.Dispatcher
	remote *mqtt.Dispatcher
}

func create(r Rule, hub, remote config.MQTTClient) *ruler {
	if r.ID != "" {
		hub.CleanSession = false
		remote.CleanSession = false
		hub.ClientID = fmt.Sprintf("%s-%s", hub.ClientID, r.ID)
		remote.ClientID = fmt.Sprintf("%s-%s", remote.ClientID, r.ID)
	} else {
		tmpid := uuid.Generate().String()
		hub.CleanSession = false
		remote.CleanSession = false
		hub.ClientID = fmt.Sprintf("%s-%s", hub.ClientID, tmpid)
		remote.ClientID = fmt.Sprintf("%s-%s", remote.ClientID, tmpid)
	}
	hub.Subscriptions = r.Hub.Subscriptions
	remote.Subscriptions = r.Remote.Subscriptions
	return &ruler{
		r:      &r,
		hub:    mqtt.NewDispatcher(hub),
		remote: mqtt.NewDispatcher(remote),
	}
}

func (rr *ruler) start() error {
	hubHandler := mqtt.Handler{
		ProcessPublish: func(p *packet.Publish) error {
			return rr.remote.Send(p)
		},
		ProcessPuback: func(p *packet.Puback) error {
			return rr.remote.Send(p)
		},
	}
	if err := rr.hub.Start(hubHandler); err != nil {
		return err
	}
	remoteHandler := mqtt.Handler{
		ProcessPublish: func(p *packet.Publish) error {
			return rr.hub.Send(p)
		},
		ProcessPuback: func(p *packet.Puback) error {
			return rr.hub.Send(p)
		},
	}
	if err := rr.remote.Start(remoteHandler); err != nil {
		return err
	}
	return nil
}

func (rr *ruler) close() {
	rr.hub.Close()
	rr.remote.Close()
}
