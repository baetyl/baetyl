package engine

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/chain"
	"github.com/baetyl/baetyl/v2/sync"
)

type handlerDownside struct {
	*engineImpl
}

func (h *handlerDownside) OnMessage(msg interface{}) error {
	m := msg.(*v1.Message)
	h.log.Debug("engine downside msg", log.Any("msg", m))

	key := fmt.Sprintf("%s_%s_%s_%s", m.Metadata["namespace"], m.Metadata["name"], m.Metadata["container"], m.Metadata["token"])
	downside := fmt.Sprintf("%s_%s", key, "down")
	h.log.Debug("engine pub downside topic", log.Any("topic", downside))

	switch m.Kind {
	case v1.MessageCMD:
		if m.Metadata["cmd"] == "connect" {
			old, ok := h.chains.Load(key)
			if ok {
				old.(chain.Chain).Close()
				h.chains.Delete(key)
				h.log.Debug("close chain", log.Any("chain name", key))
			}
			h.log.Debug("new chain", log.Any("chain name", key))
			c, err := chain.NewChain(h.cfg, h.ami, m.Metadata)
			if err != nil {
				h.pb.Publish(sync.TopicUpside, &v1.Message{
					Kind: v1.MessageCMD,
					Metadata: map[string]string{
						"success": "false",
						"msg":     "failed to connect",
						"token":   m.Metadata["token"],
					},
				})
				return err
			}
			err = c.Start()
			if err != nil {
				h.pb.Publish(sync.TopicUpside, &v1.Message{
					Kind: v1.MessageCMD,
					Metadata: map[string]string{
						"success": "false",
						"msg":     "failed to exec",
						"token":   m.Metadata["token"],
					},
				})
				return err
			}
			h.chains.Store(key, c)
		}
	case v1.MessageData:
		if _, ok := h.chains.Load(key); !ok {
			return h.pb.Publish(sync.TopicUpside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to find connect chain",
					"token":   m.Metadata["token"],
				},
			})
		}
		err := h.pb.Publish(downside, m)
		if err != nil {
			h.pb.Publish(sync.TopicUpside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to publish downside chain",
					"token":   m.Metadata["token"],
				},
			})
		}
	default:
		h.log.Warn("remote debug message kind not support", log.Any("msg", m))
	}
	return nil
}

func (h *handlerDownside) OnTimeout() error {
	return h.pb.Publish(sync.TopicUpside, &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "engine timeout",
		},
	})
}
