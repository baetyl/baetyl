package eventx

import (
	"encoding/json"
	"fmt"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	"github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/config"
)

type handler struct {
	mqtt *mqtt.Client
	log  *log.Logger
	cfg  config.EventConfig
}

func (h *handler) OnMessage(msg interface{}) error {
	message, _ := msg.(*v1.Message)
	switch message.Kind {
	case v1.MessageNodeProps:
		var propsDelta v1.Delta
		if err := message.Content.Unmarshal(&propsDelta); err != nil {
			return errors.Trace(err)
		}
		if len(propsDelta) == 0 {
			return nil
		}
		pld, err := json.Marshal(propsDelta)
		if err != nil {
			return errors.Trace(err)
		}
		if err = h.mqtt.Publish(mqtt.QOS(h.cfg.Publish.QOS),
			h.cfg.Publish.Topic, pld, 0, false, false); err != nil {
			return errors.Trace(err)
		}
		h.log.Debug("send node props to mqtt broker successfully", log.Any("props", propsDelta))
	case v1.MessageCMD:
		switch message.Metadata["cmd"] {
		case v1.MessageRPCMqtt:
			request := &v1.RPCMqttMessage{}
			err := message.Content.Unmarshal(request)
			if err != nil {
				return errors.Trace(err)
			}
			var buf []byte
			if request.Content != nil {
				buf = []byte(fmt.Sprintf("%v", request.Content))
			}
			if err = h.mqtt.Publish(mqtt.QOS(request.QoS),
				request.Topic, buf, 0, false, false); err != nil {
				return errors.Trace(err)
			}
			h.log.Debug("send rpc to mqtt broker successfully", log.Any("props", request))
		default:
			h.log.Debug("unknown command", log.Any("cmd", message.Metadata["cmd"]))
		}
	default:
		h.log.Debug("message kind not support yet", log.Any("type", message.Kind))
	}
	return nil
}

func (h *handler) OnTimeout() error {
	return nil
}
