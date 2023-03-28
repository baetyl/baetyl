package engine

import (
	"errors"
	"fmt"
	"os"
	gosync "sync"
	"testing"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/plugin"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/eventx"
	"github.com/baetyl/baetyl/v2/mock"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	errDisconnect = "disconnect"
)

var (
	engMsgWG = gosync.WaitGroup{}
)

func genHandlerDownsideEngine(t *testing.T) (*engineImpl, *mock.MockAMI, *gomock.Controller) {
	t.Setenv(context.KeySvcName, specV1.BaetylCore)
	// prepare struct
	cfg := config.Config{}

	pb, err := pubsub.NewPubsub(0)
	assert.NoError(t, err)

	plugin.RegisterFactory("defaultpubsub", func() (plugin.Plugin, error) {
		return pb, nil
	})

	ctl := gomock.NewController(t)
	ami := mock.NewMockAMI(ctl)

	e := &engineImpl{
		cfg:    cfg,
		pb:     pb,
		ami:    ami,
		chains: gosync.Map{},
		log:    log.With(),
	}

	return e, ami, ctl
}

func TestHandlerDownside(t *testing.T) {
	e, ami, ctl := genHandlerDownsideEngine(t)

	meta := map[string]string{
		"namespace": "default",
		"name":      "core",
		"container": "xxx",
		"token":     "0123456789",
		"cmd":       "connect",
	}

	h := &handlerDownside{engineImpl: e}

	// test
	ch, err := e.pb.Subscribe(sync.TopicUpside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	pro := pubsub.NewProcessor(ch, 0, &msgUpside{t: t, pb: e.pb})
	pro.Start()

	// msg 0
	msg0 := &specV1.Message{
		Kind:     specV1.MessageCMD,
		Metadata: meta,
	}
	engMsgWG.Add(1)
	err = h.OnMessage(msg0)
	assert.Error(t, err)

	// msg 1
	msg1 := &specV1.Message{
		Kind:     specV1.MessageData,
		Metadata: meta,
	}
	engMsgWG.Add(1)
	err = h.OnMessage(msg1)
	assert.Error(t, err)

	// cmd store msg 2
	key := fmt.Sprintf("%s_%s_%s_%s", meta["namespace"], meta["name"], meta["container"], meta["token"])
	mockChain := mock.NewMockChain(ctl)
	h.chains.Store(key, mockChain)
	mockChain.EXPECT().Close().Return(nil).Times(1)

	h.cfg.Plugin.Pubsub = "defaultpubsub"
	msg2 := &specV1.Message{
		Kind:     specV1.MessageCMD,
		Metadata: meta,
	}

	ami.EXPECT().RemoteCommand(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	engMsgWG.Add(1)
	err = h.OnMessage(msg2)
	assert.NoError(t, err)

	// msg 3
	msg3 := &specV1.Message{
		Kind:     specV1.MessageData,
		Metadata: meta,
	}
	err = h.OnMessage(msg3)
	assert.NoError(t, err)

	// msg 4
	msg4 := &specV1.Message{
		Kind:     specV1.MessageKeep,
		Metadata: meta,
	}
	err = h.OnMessage(msg4)
	assert.NoError(t, err)

	// msg 5
	engMsgWG.Add(1)
	err = h.OnTimeout()
	assert.NoError(t, err)

	// msg 6 disconnect not exist
	msg6 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"namespace": "default",
			"name":      "core",
			"container": "xxx",
			"token":     "0123456789",
			"cmd":       "disconnect",
		},
	}
	h.chains.Delete(key)
	err = h.OnMessage(msg6)
	assert.NoError(t, err)

	// msg 7 disconnect close error
	msg7 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"namespace": "default",
			"name":      "core",
			"container": "xxx",
			"token":     "0123456789",
			"cmd":       "disconnect",
		},
	}
	h.chains.Store(key, mockChain)
	mockChain.EXPECT().Close().Return(os.ErrInvalid).Times(1)
	engMsgWG.Add(1)
	err = h.OnMessage(msg7)
	assert.Error(t, err, os.ErrInvalid)

	engMsgWG.Wait()
}

type msgUpside struct {
	t  *testing.T
	pb pubsub.Pubsub
}

func (h *msgUpside) OnMessage(msg interface{}) error {
	m, ok := msg.(*specV1.Message)
	assert.True(h.t, ok)
	assert.Equal(h.t, "false", m.Metadata["success"])

	fmt.Println(m.Metadata["msg"])
	switch m.Metadata["msg"] {
	case ErrCreateChain, ErrGetChain, ErrTimeout, errDisconnect, ErrCloseChain:
		engMsgWG.Done()
	default:
		assert.Fail(h.t, "unexpected message")
	}
	return nil
}

func (h *msgUpside) OnTimeout() error {
	return nil
}

func TestHandlerDownsideLabels(t *testing.T) {
	e, ami, _ := genHandlerDownsideEngine(t)
	h := &handlerDownside{engineImpl: e}

	// test
	ch, err := e.pb.Subscribe(sync.TopicUpside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	handler := &msgLabelUpside{t: t}
	pro := pubsub.NewProcessor(ch, 0, handler)
	pro.Start()

	// msg 8 update node label success
	engMsgWG.Add(1)
	msg8 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd":     specV1.MessageCommandNodeLabel,
			"subName": "node01",
		},
		Content: specV1.LazyValue{
			Value: map[string]string{
				"beta.kubernetes.io/arch":        "amd64",
				"beta.kubernetes.io/os":          "linux",
				"kubernetes.io/arch":             "amd64",
				"kubernetes.io/hostname":         "docker-desktop",
				"kubernetes.io/os":               "linux",
				"node-role.kubernetes.io/master": "",
				"a":                              "b",
			},
		},
	}
	ami.EXPECT().UpdateNodeLabels("node01", gomock.Any()).Return(nil).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(msg8)
	assert.NoError(t, err)
	engMsgWG.Wait()

	// msg 9 update node label err sub name
	engMsgWG.Add(1)
	msg9 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageCommandNodeLabel,
		},
	}
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "false", m.Metadata["success"])
		assert.Equal(t, ErrSubNodeName, m.Metadata["msg"])
		engMsgWG.Done()
	}
	err = h.OnMessage(msg9)
	assert.Error(t, err, ErrSubNodeName)
	engMsgWG.Wait()

	// msg10 update multiple node label
	engMsgWG.Add(1)
	labels := map[string]string{
		"beta.kubernetes.io/arch":        "amd64",
		"beta.kubernetes.io/os":          "linux",
		"kubernetes.io/arch":             "amd64",
		"kubernetes.io/hostname":         "docker-desktop",
		"kubernetes.io/os":               "linux",
		"node-role.kubernetes.io/master": "",
		"a":                              "b",
	}
	msg10 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageCommandMultiNodeLabels,
		},
		Content: specV1.LazyValue{
			Value: map[string]map[string]string{
				"node-1": labels,
				"node-2": labels,
			},
		},
	}
	ami.EXPECT().UpdateNodeLabels("node-1", labels).Return(nil).Times(1)
	ami.EXPECT().UpdateNodeLabels("node-2", labels).Return(nil).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(msg10)
	assert.NoError(t, err)
	engMsgWG.Wait()

	// msg11 update multiple node label failed
	engMsgWG.Add(1)
	msg11 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageCommandMultiNodeLabels,
		},
		Content: specV1.LazyValue{
			Value: map[string]map[string]string{
				"node-1": labels,
			},
		},
	}
	ami.EXPECT().UpdateNodeLabels("node-1", labels).Return(errors.New("err")).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "false", m.Metadata["success"])
		assert.Equal(t, "err", m.Metadata["msg"])
		engMsgWG.Done()
	}
	err = h.OnMessage(msg11)
	assert.Error(t, err)
	engMsgWG.Wait()

	// msg12 describe
	engMsgWG.Add(1)
	msg12 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd":          specV1.MessageCommandDescribe,
			"resourceType": "po",
			"namespace":    "baetyl-edge",
			"name":         "feitian-lili-945",
		},
	}
	ami.EXPECT().RemoteDescribe(msg12.Metadata["resourceType"],
		msg12.Metadata["namespace"], msg12.Metadata["name"]).Return("desc", nil).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(msg12)
	assert.NoError(t, err)
	engMsgWG.Wait()
}

type msgLabelUpside struct {
	t     *testing.T
	check func(msg interface{})
}

func (h *msgLabelUpside) OnMessage(msg interface{}) error {
	h.check(msg)
	return nil
}

func (h *msgLabelUpside) OnTimeout() error {
	return nil
}

func TestHandlerDownsideRPC(t *testing.T) {
	e, ami, _ := genHandlerDownsideEngine(t)
	h := &handlerDownside{engineImpl: e}

	// test
	ch, err := e.pb.Subscribe(sync.TopicUpside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	handler := &msgRPCUpside{t: t}
	pro := pubsub.NewProcessor(ch, 0, handler)
	pro.Start()

	// req0 rpc success
	engMsgWG.Add(1)
	req0 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageRPC,
		},
		Content: specV1.LazyValue{
			Value: &specV1.RPCRequest{
				App:    "app",
				Method: "get",
				System: true,
				Params: "",
				Header: map[string]string{},
				Body:   "",
			},
		},
	}
	res0 := &specV1.RPCResponse{
		StatusCode: 200,
		Header:     map[string][]string{},
		Body:       []byte{},
	}
	ami.EXPECT().RPCApp("http://app.baetyl-edge-system", gomock.Any()).Return(res0, nil).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(req0)
	assert.NoError(t, err)
	engMsgWG.Wait()

	// req1 rpc success
	engMsgWG.Add(1)
	req1 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageRPC,
		},
		Content: specV1.LazyValue{
			Value: &specV1.RPCRequest{
				App:    "http://127.0.0.1",
				Method: "get",
				System: true,
				Params: "",
				Header: map[string]string{},
				Body:   "",
			},
		},
	}
	res1 := &specV1.RPCResponse{
		StatusCode: 200,
		Header:     map[string][]string{},
		Body:       []byte{},
	}
	ami.EXPECT().RPCApp("http://127.0.0.1", gomock.Any()).Return(res1, nil).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(req1)
	assert.NoError(t, err)
	engMsgWG.Wait()

	// req2 rpc fail
	engMsgWG.Add(1)
	req2 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageRPC,
		},
		Content: specV1.LazyValue{
			Value: &specV1.RPCRequest{
				App:    "app",
				Method: "get",
				System: false,
				Params: "",
				Header: map[string]string{},
				Body:   "",
			},
		},
	}
	ami.EXPECT().RPCApp("http://app.baetyl-edge", gomock.Any()).Return(nil, errors.New("timeout")).Times(1)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "false", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(req2)
	assert.NotNil(t, err)
	engMsgWG.Wait()
}

func TestHandlerDownsideAgent(t *testing.T) {
	e, _, ctl := genHandlerDownsideEngine(t)
	h := &handlerDownside{engineImpl: e}

	agentClient := mock.NewMockAgentClient(ctl)
	h.agentClient = agentClient

	// test
	ch, err := e.pb.Subscribe(sync.TopicUpside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	handler := &msgRPCUpside{t: t}
	pro := pubsub.NewProcessor(ch, 0, handler)
	pro.Start()

	// req0 set success
	engMsgWG.Add(1)
	req0 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd":    specV1.MessageAgent,
			"action": "open",
		},
		Content: specV1.LazyValue{},
	}
	agentClient.EXPECT().GetOrSetAgentFlag("open").Return(true, nil)
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(req0)
	assert.NoError(t, err)
	engMsgWG.Wait()

	// req1 set fail
	engMsgWG.Add(1)
	req1 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd":    specV1.MessageAgent,
			"action": "close",
		},
		Content: specV1.LazyValue{},
	}
	agentClient.EXPECT().GetOrSetAgentFlag("close").Return(false, errors.New("timeout"))
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "false", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(req1)
	assert.NotNil(t, err)
	engMsgWG.Wait()

	// req2 set fail
	engMsgWG.Add(1)
	req2 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd":    specV1.MessageAgent,
			"action": "close",
		},
		Content: specV1.LazyValue{},
	}
	h.agentClient = nil
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "false", m.Metadata["success"])
		engMsgWG.Done()
	}
	err = h.OnMessage(req2)
	assert.NotNil(t, err)
	engMsgWG.Wait()
}

type msgRPCUpside struct {
	t     *testing.T
	check func(msg interface{})
}

func (h *msgRPCUpside) OnMessage(msg interface{}) error {
	h.check(msg)
	return nil
}

func (h *msgRPCUpside) OnTimeout() error {
	return nil
}

func TestHandlerDownsideMqtt(t *testing.T) {
	e, _, _ := genHandlerDownsideEngine(t)
	h := &handlerDownside{engineImpl: e}

	// upside msg
	ch, err := e.pb.Subscribe(sync.TopicUpside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	handler := &msgRPCUpside{t: t}
	pro := pubsub.NewProcessor(ch, 0, handler)
	pro.Start()
	defer pro.Close()

	// event msg
	ch0, err := e.pb.Subscribe(eventx.TopicEvent)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	handler0 := &msgRPCMqtt{t: t}
	pro0 := pubsub.NewProcessor(ch0, 0, handler0)
	pro0.Start()
	defer pro0.Close()

	// req0 rpc success
	engMsgWG.Add(2)
	req0 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageRPCMqtt,
		},
		Content: specV1.LazyValue{
			Value: &specV1.RPCMqttMessage{
				Topic:   "test/node",
				QoS:     0,
				Content: "result",
			},
		},
	}
	handler.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		assert.Equal(t, "true", m.Metadata["success"])
		engMsgWG.Done()
	}
	handler0.check = func(msg interface{}) {
		m, ok := msg.(*specV1.Message)
		assert.True(t, ok)
		request := &specV1.RPCMqttMessage{}
		handlerErr := m.Content.Unmarshal(request)
		assert.NoError(t, handlerErr)
		var buf []byte
		if request.Content != nil {
			buf = []byte(fmt.Sprintf("%v", request.Content))
		}
		assert.Equal(t, string(buf), "result")
		engMsgWG.Done()
	}
	err = h.OnMessage(req0)
	assert.NoError(t, err)
	engMsgWG.Wait()
}

type msgRPCMqtt struct {
	t     *testing.T
	check func(msg interface{})
}

func (h *msgRPCMqtt) OnMessage(msg interface{}) error {
	h.check(msg)
	return nil
}

func (h *msgRPCMqtt) OnTimeout() error {
	return nil
}
