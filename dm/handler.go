// Package dm 设备管理实现
package dm

import (
	"encoding/json"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	"github.com/baetyl/baetyl-go/v2/spec/v1"
)

var (
	ErrInvalidMessage = errors.New("failed to parse message")
)

type handler struct {
	msgCh chan *v1.Message
	log   *log.Logger
}

func newHandler(msgCh chan *v1.Message, log *log.Logger) pubsub.Handler {
	return &handler{
		msgCh: msgCh,
		log:   log,
	}
}

func (h *handler) OnMessage(msg interface{}) error {
	message, ok := msg.(*v1.Message)
	if !ok {
		h.log.Error("invalid message", log.Any("message", msg))
		return errors.Trace(ErrInvalidMessage)
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		h.log.Error("failed to marshal msg")
	}
	msgStr := string(bytes)
	switch message.Kind {
	case v1.MessageDevices, v1.MessageDeviceEvent, v1.MessageDevicePropertyGet:
		select {
		case h.msgCh <- message:
			h.log.Debug("handler receive upside message", log.Any("msg", msgStr))
		default:
			h.log.Error("failed to write device message to channel", log.Any("msg", msgStr))
		}
	case v1.MessageDeviceDelta:
		h.log.Warn("set number has been implemented by sending mqtt directly")
	default:
		h.log.Debug("message kind not support yet", log.Any("type", message.Kind))
	}
	return nil
}

func (h *handler) OnTimeout() error {
	return nil
}
