package chain

import (
	"bytes"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/sync"
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
			errPub := h.pb.Publish(h.upside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to unmarshal data message",
					"token":   h.token,
				},
			})
			if errPub != nil {
				h.log.Error("failed to publish unmarshal message", log.Error(err))
			}
			return errors.Trace(err)
		}
		if bytes.Equal([]byte(ExitCmd), cmd) {
			return h.onExitMessage()
		}
		_, err = h.pipe.InWriter.Write(cmd)
		if err != nil {
			h.log.Error("failed to write debug command", log.Error(err))
			errPub := h.pb.Publish(h.upside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to write debug command",
					"token":   h.token,
				},
			})
			if errPub != nil {
				h.log.Error("failed to publish debug message", log.Error(err))
			}
			return errors.Trace(err)
		}
	default:
		h.log.Warn("remote debug message kind not support", log.Any("msg", m))
	}
	return nil
}

// onExitMessage call cancel function before close
func (h *chainHandler) onExitMessage() error {
	return h.Cancel()
}

func (h *chainHandler) OnTimeout() error {
	err := h.pb.Publish(h.upside, &v1.Message{
		Kind: v1.MessageData,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "chain timeout",
			"token":   h.token,
		},
	})
	if err != nil {
		h.log.Error("failed to publish timeout message", log.Error(err))
	}
	return h.pb.Publish(sync.TopicDownside, &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"namespace": h.debugOptions.Namespace,
			"name":      h.debugOptions.Name,
			"container": h.debugOptions.Container,
			"token":     h.token,
			"cmd":       "disconnect",
		},
	})
}
