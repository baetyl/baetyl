package engine

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/chain"
	"github.com/baetyl/baetyl/v2/eventx"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	ErrCreateChain          = "failed to create new chain"
	ErrCloseChain           = "failed to close connected chain"
	ErrGetChain             = "failed to get connected chain"
	ErrPublishDownsideChain = "failed to publish downside chain"
	ErrExecData             = "failed to exec"
	ErrSubNodeName          = "failed to get sub node name"
	ErrAgentNotSet          = "failed to request for the not set agent"
	ErrTimeout              = "engine timeout"

	PrefixHTTP  = "http://"
	PrefixHTTPS = "https://"
)

type handlerDownside struct {
	*engineImpl
}

func (h *handlerDownside) OnMessage(msg interface{}) error {
	m := msg.(*v1.Message)
	h.log.Debug("engine downside msg", log.Any("msg", m))

	// Todo : improve, only the core module supports remote debugging
	if os.Getenv(context.KeySvcName) != v1.BaetylCore {
		return nil
	}

	key := fmt.Sprintf("%s_%s_%s_%s", m.Metadata["namespace"], m.Metadata["name"], m.Metadata["container"], m.Metadata["token"])
	downside := fmt.Sprintf("%s_%s", key, "down")
	h.log.Debug("engine pub downside topic", log.Any("topic", downside))

	switch m.Kind {
	case v1.MessageCMD:
		switch m.Metadata["cmd"] {
		case v1.MessageCommandConnect:
			err := h.connect(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageCommandLogs:
			err := h.viewLogs(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageCommandDisconnect:
			err := h.disconnect(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageCommandNodeLabel:
			err := h.nodeLabel(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageCommandMultiNodeLabels:
			err := h.labelMultiNodes(key, m)
			if err != nil {
				return err
			}
		case v1.MessageRPC:
			err := h.rpc(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageAgent:
			err := h.agentControl(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageRPCMqtt:
			err := h.rpcMqtt(key, m)
			if err != nil {
				return errors.Trace(err)
			}
		case v1.MessageCommandDescribe:
			err := h.describe(key, m)
			if err != nil {
				return errors.Trace(err)
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
			return errors.Trace(err)
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

func (h *handlerDownside) viewLogs(key string, m *v1.Message) error {
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

	opt := &ami.LogsOptions{}
	err := m.Content.Unmarshal(&opt)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}

	// create new chain
	c, err := chain.NewChain(h.cfg, h.ami, m.Metadata, false)
	if err != nil {
		h.publishFailedMsg(key, ErrCreateChain, m)
		return errors.Trace(err)
	}
	err = c.ViewLogs(opt)
	if err != nil {
		h.publishFailedMsg(key, ErrExecData, m)
		return errors.Trace(err)
	}
	h.chains.Store(key, c)
	return nil
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
	c, err := chain.NewChain(h.cfg, h.ami, m.Metadata, true)
	if err != nil {
		h.publishFailedMsg(key, ErrCreateChain, m)
		return errors.Trace(err)
	}
	err = c.Debug()
	if err != nil {
		h.publishFailedMsg(key, ErrExecData, m)
		return errors.Trace(err)
	}
	h.chains.Store(key, c)
	return nil
}

func (h *handlerDownside) sendExit(key string) {
	m := &v1.Message{
		Kind:    v1.MessageData,
		Content: v1.LazyValue{Value: []byte(chain.ExitCmd)},
	}
	downside := fmt.Sprintf("%s_%s", key, "down")
	err := h.pb.Publish(downside, m)
	if err != nil {
		h.log.Error(ErrPublishDownsideChain, log.Error(errors.Trace(err)))
	}
}

func (h *handlerDownside) disconnect(key string, m *v1.Message) error {
	c, ok := h.chains.Load(key)
	if !ok {
		return nil
	}
	h.sendExit(key)
	err := c.(chain.Chain).Close()
	if err != nil {
		h.publishFailedMsg(key, ErrCloseChain, m)
		return errors.Trace(err)
	}
	h.chains.Delete(key)
	return nil
}

func (h *handlerDownside) nodeLabel(key string, m *v1.Message) error {
	nodeName, ok := m.Metadata["subName"]
	if !ok {
		h.publishFailedMsg(key, ErrSubNodeName, m)
		return errors.New(ErrSubNodeName)
	}
	labels := new(map[string]string)
	err := m.Content.Unmarshal(labels)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	err = h.ami.UpdateNodeLabels(nodeName, *labels)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	h.publishSuccessMsg(key, m)
	return nil
}

func (h *handlerDownside) labelMultiNodes(key string, m *v1.Message) error {
	var nodesLabels map[string]map[string]string
	err := m.Content.Unmarshal(&nodesLabels)
	if err != nil {
		return errors.Trace(err)
	}

	var errs []string
	for name, labels := range nodesLabels {
		err = h.ami.UpdateNodeLabels(name, labels)
		if err != nil {
			h.log.Warn(err.Error())
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		es := strings.Join(errs, "\n")
		h.log.Warn(es)
		h.publishFailedMsg(key, es, m)
		return errors.Trace(errors.New(es))
	}
	h.publishSuccessMsg(key, m)
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

func (h *handlerDownside) publishSuccessMsg(key string, m *v1.Message) {
	errPublish := h.pb.Publish(sync.TopicUpside, &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "true",
			"token":   m.Metadata["token"],
		},
	})
	if errPublish != nil {
		h.log.Error("failed to publish message", log.Any("topic", sync.TopicUpside), log.Any("chain name", key), log.Error(errPublish))
	}
}

func (h *handlerDownside) rpc(key string, m *v1.Message) error {
	request := &v1.RPCRequest{}
	err := m.Content.Unmarshal(request)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	token := m.Metadata["token"]
	res, err := h.ami.RPCApp(assembleUrl(request), request)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	h.log.Debug("rpc success", log.Any("status", res.StatusCode))
	response := &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "true",
			"token":   token,
		},
		Content: v1.LazyValue{
			Value: res,
		},
	}
	err = h.pb.Publish(sync.TopicUpside, response)
	if err != nil {
		h.log.Error("failed to publish message", log.Any("topic", sync.TopicUpside), log.Any("chain name", key), log.Error(err))
	}
	return nil
}

func (h *handlerDownside) rpcMqtt(key string, m *v1.Message) error {
	token := m.Metadata["token"]
	err := h.pb.Publish(eventx.TopicEvent, m)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	h.log.Debug("rpc success", log.Any("key", key))
	response := &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "true",
			"token":   token,
		},
		Content: v1.LazyValue{},
	}
	err = h.pb.Publish(sync.TopicUpside, response)
	if err != nil {
		h.log.Error("failed to publish message", log.Any("topic", sync.TopicUpside), log.Any("chain name", key), log.Error(err))
	}
	return nil
}

func (h *handlerDownside) agentControl(key string, m *v1.Message) error {
	if h.agentClient == nil {
		h.publishFailedMsg(key, ErrAgentNotSet, m)
		return errors.Trace(errors.New(ErrAgentNotSet))
	}
	res, err := h.agentClient.GetOrSetAgentFlag(m.Metadata["action"])
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	h.log.Debug("agent info", log.Any("status", res))
	response := &v1.Message{
		Kind: v1.MessageCMD,
		Metadata: map[string]string{
			"success": "true",
			"token":   m.Metadata["token"],
			"stat":    strconv.FormatBool(res),
		},
		Content: v1.LazyValue{},
	}
	err = h.pb.Publish(sync.TopicUpside, response)
	if err != nil {
		h.log.Error("failed to publish message", log.Any("topic", sync.TopicUpside), log.Any("chain name", key), log.Error(err))
	}
	return nil
}

func (h *handlerDownside) describe(key string, m *v1.Message) error {
	ns, n := m.Metadata["namespace"], m.Metadata["name"]
	resourceType := m.Metadata["resourceType"]
	res, err := h.ami.RemoteDescribe(resourceType, ns, n)
	if err != nil {
		h.publishFailedMsg(key, err.Error(), m)
		return errors.Trace(err)
	}
	h.log.Debug("describe pod success", log.Any("key", key))
	response := &v1.Message{
		Kind: v1.MessageData,
		Metadata: map[string]string{
			"success": "true",
			"token":   m.Metadata["token"],
		},
		Content: v1.LazyValue{Value: []byte(res)},
	}
	err = h.pb.Publish(sync.TopicUpside, response)
	if err != nil {
		h.log.Error("failed to publish message", log.Any("topic", sync.TopicUpside), log.Any("chain name", key), log.Error(err))
	}
	return nil
}

func assembleUrl(req *v1.RPCRequest) string {
	url := req.App
	if !strings.Contains(url, PrefixHTTP) && !strings.Contains(url, PrefixHTTPS) {
		if req.System {
			url = PrefixHTTP + url + ".baetyl-edge-system"
		} else {
			url = PrefixHTTP + url + ".baetyl-edge"
		}
	}
	url += req.Params
	return url
}
