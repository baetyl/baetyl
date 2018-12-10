package main

import (
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/function"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/docker/distribution/uuid"
	"github.com/juju/errors"
)

type ruler struct {
	r  *function.Rule
	md *mqtt.Dispatcher
	fd *function.Dispatcher
}

func create(r function.Rule, cc mqtt.ClientConfig, f *function.Function) (*ruler, error) {
	if r.ID != "" {
		cc.CleanSession = false
		cc.ClientID = fmt.Sprintf("%s-%s", cc.ClientID, r.ID)
	} else {
		cc.CleanSession = true
		cc.ClientID = fmt.Sprintf("%s-%s", cc.ClientID, uuid.Generate().String())
	}
	cc.Subscriptions = []mqtt.Subscription{mqtt.Subscription{Topic: r.Subscribe.Topic, QOS: r.Subscribe.QOS}}
	fd, err := function.NewDispatcher(f)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &ruler{
		r:  &r,
		fd: fd,
		md: mqtt.NewDispatcher(cc),
	}, nil
}

func (rr *ruler) start() error {
	rr.fd.SetCallback(func(pkt *packet.Publish) {
		subqos := pkt.Message.QOS
		if pkt.Message.Payload != nil {
			if pkt.Message.QOS > rr.r.Publish.QOS {
				pkt.Message.QOS = rr.r.Publish.QOS
			}
			pkt.Message.Topic = rr.r.Publish.Topic
			err := rr.md.Send(pkt)
			if err != nil {
				return
			}
		}
		if subqos == 1 && (rr.r.Publish.QOS == 0 || pkt.Message.Payload == nil) {
			puback := packet.NewPuback()
			puback.ID = pkt.ID
			rr.md.Send(puback)
		}
	})
	err := rr.md.Start(func(pkt packet.Generic) {
		switch p := pkt.(type) {
		case *packet.Publish:
			rr.fd.Invoke(p)
		case *packet.Puback:
			rr.md.Send(p)
		}
	})
	return errors.Trace(err)
}

func (rr *ruler) close() {
	rr.md.Close()
	rr.fd.Close()
}
