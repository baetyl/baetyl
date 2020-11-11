package eventx

import (
	"encoding/json"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/node"
)

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
		if props, ok := delta[node.KeyNodeProps]; ok && props != nil {
			pld, err := json.Marshal(props)
			if err != nil {
				return err
			}
			err = h.mqtt.Publish(DefaultMQTTQOS, NodePropsTopic, pld, 0, false, false)
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
