package main

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/sdk-go/openedge"
)

type ruler struct {
	cfg RuleInfo
	fcd *Dispatcher
	hub *mqtt.Dispatcher
	log logger.Logger
}

func create(ctx openedge.Context, ri RuleInfo, fi FunctionInfo) *ruler {
	hub := ctx.Config().Hub
	hub.ClientID = ri.ClientID
	hub.Subscriptions = []mqtt.TopicInfo{ri.Subscribe}
	log := logger.WithField("rule", ri.ClientID)
	return &ruler{
		cfg: ri,
		hub: mqtt.NewDispatcher(hub, log),
		fcd: NewDispatcher(ctx, fi),
	}
}

func (rr *ruler) start() error {
	rr.fcd.SetCallback(func(pkt *packet.Publish) {
		subqos := byte(pkt.Message.QOS)
		if pkt.Message.Payload != nil {
			if subqos > rr.cfg.Publish.QOS {
				pkt.Message.QOS = packet.QOS(rr.cfg.Publish.QOS)
			}
			pkt.Message.Topic = rr.cfg.Publish.Topic
			err := rr.hub.Send(pkt)
			if err != nil {
				return
			}
		}
		if subqos == 1 && (rr.cfg.Publish.QOS == 0 || pkt.Message.Payload == nil) {
			puback := packet.NewPuback()
			puback.ID = pkt.ID
			rr.hub.Send(puback)
		}
	})
	return rr.hub.Start(rr)
}

func (rr *ruler) ProcessPublish(pkt *packet.Publish) error {
	return rr.fcd.Call(pkt)
}

func (rr *ruler) ProcessPuback(pkt *packet.Puback) error {
	return rr.hub.Send(pkt)
}

func (rr *ruler) ProcessError(err error) {
	rr.log.Errorf(err.Error())
}

func (rr *ruler) close() {
	rr.hub.Close()
	rr.fcd.Close()
}
