package sync

import (
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/plugin"
)

type handler struct {
	link plugin.Link
}

func (h *handler) OnMessage(msg interface{}) error {
	return h.link.Send(msg.(*v1.Message))
}

func (h *handler) OnTimeout() error {
	msg := &v1.Message{
		Kind: v1.MessageData,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "sync timeout",
		},
	}
	return h.link.Send(msg)
}
