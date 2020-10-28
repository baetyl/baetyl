package engine

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/chain"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	ErrCreateChain          = "failed to create new chain"
	ErrCloseChain           = "failed to close connected chain"
	ErrGetChain             = "failed to get connected chain"
	ErrPublishDownsideChain = "failed to publish downside chain"
	ErrExecData             = "failed to exec"
	ErrTimeout              = "engine timeout"
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
		switch m.Metadata["cmd"] {
		case "connect":
			err := h.connect(key, m)
			if err != nil {
				return err
			}
		case "disconnect":
			err := h.disconnect(key, m)
			if err != nil {
				return err
			}
		default:
			h.log.Debug("unknown command", log.Any("cmd", m.Metadata["cmd"]))
		}
	case v1.MessageData:
		if _, ok := h.chains.Load(key); !ok {
			h.publishFailedMsg(key, ErrGetChain, m)
			return errors.New(ErrGetChain + key)
		}
		err := h.pb.Publish(downside, m)
		if err != nil {
			h.log.Error(ErrPublishDownsideChain, log.Error(errors.Trace(err)))
			h.publishFailedMsg(key, ErrPublishDownsideChain, m)
			return err
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
			"msg":     ErrTimeout,
		},
	})
}

func (h *handlerDownside) connect(key string, m *v1.Message) error {
	// close old chain if exist
	old, ok := h.chains.Load(key)
	if ok {
		err := old.(chain.Chain).Close()
		if err != nil {
			h.log.Warn("failed to close old chain", log.Any("chain", key))
		}
		h.chains.Delete(key)
		h.log.Debug("close chain", log.Any("chain name", key))
	}
	h.log.Debug("new chain", log.Any("chain name", key))

	// create new chain
	c, err := chain.NewChain(h.cfg, h.ami, m.Metadata)
	if err != nil {
		h.publishFailedMsg(key, ErrCreateChain, m)
		return err
	}
	err = c.Start()
	if err != nil {
		h.publishFailedMsg(key, ErrExecData, m)
		return err
	}
	h.chains.Store(key, c)
	return nil
}

func (h *handlerDownside) disconnect(key string, m *v1.Message) error {
	c, ok := h.chains.Load(key)
	if !ok {
		return nil
	}
	err := c.(chain.Chain).Close()
	if err != nil {
		h.publishFailedMsg(key, ErrCloseChain, m)
		return err
	}
	h.chains.Delete(key)
	return nil
}

func (h *handlerDownside) publishFailedMsg(key, reason string, m *v1.Message) {
	errPublish := h.pb.Publish(sync.TopicUpside, &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "false",
			"msg":     reason,
			"token":   m.Metadata["token"],
		},
	})
	if errPublish != nil {
		h.log.Error("failed to publish message", log.Any("topic", sync.TopicUpside), log.Any("chain name", key), log.Error(errPublish))
	}
}
