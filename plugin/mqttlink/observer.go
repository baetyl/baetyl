// Package mqttlink 端云链接 mqtt 实现
package mqttlink

import (
	"encoding/json"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type observer struct {
	ch  chan *v1.Message
	log *log.Logger
}

func newObserver(ch chan *v1.Message, log *log.Logger) mqtt.Observer {
	return &observer{
		ch:  ch,
		log: log,
	}
}

func (o *observer) OnPublish(pkt *packet.Publish) error {
	msg := v1.Message{Metadata: make(map[string]string)}
	if err := json.Unmarshal(pkt.Message.Payload, &msg); err != nil {
		o.log.Error("failed to parse message", log.Error(err))
		return nil
	}
	select {
	case o.ch <- &msg:
		o.log.Debug("observer receive downside message", log.Any("msg", msg))
	default:
		o.log.Error("failed to write downside message to channel", log.Any("msg", msg))
	}
	return nil
}

func (o *observer) OnPuback(_ *packet.Puback) error {
	return nil
}

func (o *observer) OnError(err error) {
	o.log.Error("receive mqtt message error", log.Error(err))
}
