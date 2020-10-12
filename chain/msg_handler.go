package chain

import (
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

type chainHandler struct {
	*chain
}

func (h *chainHandler) OnMessage(msg interface{}) error {
	h.log.Debug("chain downside msg", log.Any("msg", msg))
	m := msg.(*v1.Message)
	switch m.Kind {
	case v1.MessageData:
		var cmd []byte
		err := m.Content.Unmarshal(&cmd)
		if err != nil {
			h.log.Error("failed to unmarshal data message", log.Error(err))
			h.pb.Publish(h.upside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to unmarshal data message",
					"token":   h.data["token"],
				},
			})
			return err
		}

		_, err = h.pipe.InWriter.Write(cmd)
		if err != nil {
			h.log.Error("failed to write debug command", log.Error(err))
			h.pb.Publish(h.upside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to write debug command",
					"token":   h.data["token"],
				},
			})
			return err
		}
	default:
		h.log.Warn("remote debug message kind not support", log.Any("msg", m))
	}
	return nil
}

func (h *chainHandler) OnTimeout() error {
	h.pb.Publish(h.upside, &v1.Message{
		Kind: v1.MessageData,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "chain timeout",
			"token":   h.data["token"],
		},
	})
	return nil
}
