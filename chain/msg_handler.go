package chain

import (
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type handlerDownside struct {
	*chainImpl
}

func (h *handlerDownside) OnMessage(msg interface{}) error {
	m := msg.(*v1.Message)
	switch m.Kind {
	case v1.MessageCMD:
		if m.Metadata["cmd"] == "connect" {
			h.tomb.Go(h.connecting)
			return h.mq.Publish(h.upside, &v1.Message{
				Kind: v1.MessageCMD,
				Metadata: map[string]string{
					"success": "true",
					"msg":     "connect success",
				},
				Content: v1.LazyValue{},
			})
		}
	case v1.MessageData:
		cmd := []byte{}
		err := m.Content.Unmarshal(&cmd)
		if err != nil {
			h.log.Error("failed to unmarshal data message", log.Error(err))
			return h.mq.Publish(h.upside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to unmarshal data message",
				},
				Content: v1.LazyValue{},
			})
		}

		_, err = h.pipe.inWriter.Write(cmd)
		if err != nil {
			h.log.Error("failed to write debug command", log.Error(err))
			return h.mq.Publish(h.upside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to write debug command",
				},
				Content: v1.LazyValue{},
			})
		}
	default:
		h.log.Warn("remote debug message kind not support", log.Any("msg", m))
	}
	return nil
}

func (h *handlerDownside) OnTimeout() error {
	return h.mq.Publish(h.upside, &v1.Message{
		Kind: v1.MessageData,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "timeout",
		},
		Content: v1.LazyValue{
			Value: []byte("timeout"),
		},
	})
}
