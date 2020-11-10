package eventx

import (
	"encoding/json"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	goplugin "github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/plugin"
)

const (
	TopicEvent     = "event"
	DefaultMQTTQOS = 0
	NodePropsTopic = "baetyl/node/props"
	TypeDelta      = "delta"
)

type Event struct {
	Type    string
	Payload interface{}
}

type EventX interface {
	Start()
	Close() error
}

type eventX struct {
	pb        plugin.Pubsub
	mqtt      *mqtt.Client
	log       *log.Logger
	processor pubsub.Processor
}

func NewEventX(cfg config.Config, ctx context.Context) (EventX, error) {
	pl, err := goplugin.GetPlugin(cfg.Plugin.Pubsub)
	if err != nil {
		return nil, errors.Trace(err)
	}
	mqtt, err := ctx.NewSystemBrokerClient(nil)
	if err != nil {
		return nil, err
	}
	pb := pl.(plugin.Pubsub)
	ch, err := pb.Subscribe(TopicEvent)
	if err != nil {
		return nil, err
	}
	log := log.With(log.Any("core", "sync"))
	processor := pubsub.NewProcessor(ch, 0, &handler{mqtt: mqtt, log: log})
	evt := &eventX{
		pb:        pb,
		mqtt:      mqtt,
		processor: processor,
		log:       log,
	}
	return evt, nil
}

func (e *eventX) Start() {
	obs := &observer{log: e.log}
	if err := e.mqtt.Start(obs); err != nil {
		e.log.Error("failed to start mqtt client", log.Error(err))
	}
	e.processor.Start()
}

func (e *eventX) Close() error {
	e.processor.Close()
	return e.mqtt.Close()
}

type handler struct {
	mqtt *mqtt.Client
	log  *log.Logger
}

func (h *handler) OnMessage(msg interface{}) error {
	event, ok := msg.(Event)
	if !ok {
		h.log.Error("received invalid message")
		return nil
	}
	switch event.Type {
	case TypeDelta:
		delta, ok := event.Payload.(specv1.Delta)
		if !ok {
			h.log.Error("message contains invalid delta")
			return nil
		}
		if len(delta) == 0 {
			return nil
		}
		if props, ok := delta["nodeProps"]; ok && props != nil {
			pld, err := json.Marshal(props)
			if err != nil {
				return err
			}
			pkt := packet.NewPublish()
			pkt.Message.Topic = NodePropsTopic
			pkt.Message.QOS = DefaultMQTTQOS
			pkt.Message.Payload = pld
			err = h.mqtt.Send(pkt)
			if err != nil {
				return err
			}
			h.log.Debug("send props to mqtt broker successfully", log.Any("props", props))
		}
	default:
		h.log.Debug("event type not support yet", log.Any("type", event.Type))
	}
	return nil
}

func (h *handler) OnTimeout() error {
	return nil
}

type observer struct {
	log *log.Logger
}

func (o *observer) OnPublish(*packet.Publish) error {
	return nil
}

func (o *observer) OnPuback(*packet.Puback) error {
	return nil
}

func (o *observer) OnError(err error) {
	o.log.Error("error occurs", log.Error(err))
}
