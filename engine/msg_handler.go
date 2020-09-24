package engine

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/chain"
	"github.com/baetyl/baetyl/v2/helper"
)

type handlerUpside struct {
	*engineImpl
}

func (h *handlerUpside) OnMessage(msg interface{}) error {
	return h.hp.Publish(helper.TopicUpside, msg)
}

func (h *handlerUpside) OnTimeout() error {
	return h.hp.Publish(helper.TopicUpside, &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "failed to find connect chain",
		},
		Content: v1.LazyValue{},
	})
}

type handlerDownside struct {
	*engineImpl
}

func (h *handlerDownside) OnMessage(msg interface{}) error {
	m := msg.(*v1.Message)
	key := fmt.Sprintf("%s_%s_%s", m.Metadata["namespace"], m.Metadata["name"], m.Metadata["container"])
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
				return h.hp.Publish(helper.TopicUpside, &v1.Message{
					Kind: v1.MessageCMD,
					Metadata: map[string]string{
						"success": "false",
						"msg":     "failed to connect",
					},
					Content: v1.LazyValue{},
				})
			}
			c.Subscribe(&handlerUpside{engineImpl: h.engineImpl})
			c.Publish(m)
			h.chains.Store(key, c)
		}
	case v1.MessageData:
		c, ok := h.chains.Load(key)
		if !ok {
			return h.hp.Publish(helper.TopicUpside, &v1.Message{
				Kind: v1.MessageData,
				Metadata: map[string]string{
					"success": "false",
					"msg":     "failed to find connect chain",
				},
			})
		}
		return c.(chain.Chain).Publish(m)
	default:
		h.log.Warn("remote debug message kind not support", log.Any("msg", m))
	}
	return nil
}

func (h *handlerDownside) OnTimeout() error {
	return h.hp.Publish(helper.TopicUpside, &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "false",
			"msg":     "timeout",
		},
	})
}
