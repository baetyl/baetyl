package main

import (
	"fmt"
	"sync"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// Ruler struct
type Ruler struct {
	cfg RuleInfo
	cli Client
	hub *mqtt.Dispatcher
	log logger.Logger
	tm  sync.Map
}

// NewRuler can create a ruler
func NewRuler(ctx baetyl.Context, rule RuleInfo, cli Client) *Ruler {
	hub := ctx.Config().Hub
	hub.ClientID = rule.ClientID
	hub.Subscriptions = []mqtt.TopicInfo{rule.Subscribe}
	log := logger.WithField("rule", rule.ClientID)
	return &Ruler{
		cfg: rule,
		hub: mqtt.NewDispatcher(hub, log),
		cli: cli,
		log: log,
	}
}

// Start can create a ruler
func (r *Ruler) Start() error {
	return r.hub.Start(r)
}

// Close can create a ruler
func (r *Ruler) Close() {
	r.hub.Close()
}

// ProcessPublish can create a ruler
func (r *Ruler) ProcessPublish(pkt *packet.Publish) error {
	event, err := r.processEvent(pkt)
	if err != nil {
		r.log.Errorf(err.Error())
		return err
	}
	msg := &EventMessage{
		ID:    uint64(pkt.ID),
		QOS:   uint32(pkt.Message.QOS),
		Topic: pkt.Message.Topic,
		Event: event,
	}
	return r.RuleHandler(msg)
}

func (r *Ruler) processEvent(pkt *packet.Publish) (*Event, error) {
	r.log.Debugln("event: ", string(pkt.Message.Payload))
	e, err := NewEvent(pkt.Message.Payload)
	if err != nil {
		return nil, fmt.Errorf("event invalid: %s", err.Error())
	}
	return e, nil
}

// ProcessPuback test
func (r *Ruler) ProcessPuback(pkt *packet.Puback) error {
	return nil
}

// ProcessError test
func (r *Ruler) ProcessError(err error) {
	r.log.Errorf(err.Error())
}

// RuleHandler filter topic & handler
func (r *Ruler) RuleHandler(msg *EventMessage) error {
	if msg.QOS == 1 {
		if _, ok := r.tm.Load(msg.ID); !ok {
			r.tm.Store(msg.ID, struct{}{})
		} else {
			return nil
		}
	}
	return r.cli.CallAsync(msg, r.callback)
}

func (r *Ruler) callback(msg *EventMessage, err error) {
	if msg.QOS == 1 {
		if err == nil {
			puback := packet.NewPuback()
			puback.ID = packet.ID(msg.ID)
			r.hub.Send(puback)
		}
		r.tm.Delete(msg.ID)
	}
	if err != nil {
		r.log.Errorf(err.Error())
	}
}
