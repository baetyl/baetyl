package main

import (
	"encoding/json"
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/docker/distribution/uuid"
)

type ruler struct {
	cfg RuleInfo
	fun *Function
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
		fun: NewFunction(fi, newProducer(ctx, fi)),
		log: log,
	}
}

func (rr *ruler) start() error {
	return rr.hub.Start(rr)
}

func (rr *ruler) ProcessPublish(pkt *packet.Publish) error {
	msg := &openedge.FunctionMessage{
		ID:               uint64(pkt.ID),
		QOS:              uint32(pkt.Message.QOS),
		Topic:            pkt.Message.Topic,
		Payload:          pkt.Message.Payload,
		FunctionName:     rr.cfg.Function.Name,
		FunctionInvokeID: uuid.Generate().String(),
	}
	return rr.fun.CallAsync(msg, rr.callback)
}

func (rr *ruler) ProcessPuback(pkt *packet.Puback) error {
	return rr.hub.Send(pkt)
}

func (rr *ruler) ProcessError(err error) {
	rr.log.Errorf(err.Error())
}

func (rr *ruler) close() {
	rr.hub.Close()
	rr.fun.Close()
}

func (rr *ruler) callback(in, out *openedge.FunctionMessage, err error) {
	if err != nil {
		for index := 1; index < rr.cfg.Retry.Max && err != nil; index++ {
			rr.log.Debugf("retry %d", index)
			out, err = rr.fun.Call(in)
		}
	}
	pkt := packet.NewPublish()
	pkt.ID = packet.ID(in.ID)
	pkt.Message.QOS = packet.QOS(rr.cfg.Publish.QOS)
	if in.QOS < rr.cfg.Publish.QOS {
		pkt.Message.QOS = packet.QOS(in.QOS)
	}
	pkt.Message.Topic = rr.cfg.Publish.Topic
	if err != nil {
		s := map[string]interface{}{
			"functionMessage": in,
			"errorMessage":    err.Error(),
			"errorType":       fmt.Sprintf("%T", err),
		}
		pkt.Message.Payload, _ = json.Marshal(s)
	} else if out.Payload != nil {
		pkt.Message.Payload = out.Payload
	}
	// filter
	if pkt.Message.Payload != nil {
		err := rr.hub.Send(pkt)
		if err != nil {
			return
		}
	}
	if in.QOS == 1 && (pkt.Message.QOS == 0 || pkt.Message.Payload == nil) {
		puback := packet.NewPuback()
		puback.ID = packet.ID(in.ID)
		rr.hub.Send(puback)
	}
}
