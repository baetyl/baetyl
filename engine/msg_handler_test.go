package engine

import (
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
	"github.com/baetyl/baetyl/v2/mock"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	errDisconnect = "disconnect"
)

var (
	engMsgWG = gosync.WaitGroup{}
)

func TestHandlerDownside(t *testing.T) {
	err := os.Setenv(context.KeySvcName, specV1.BaetylCore)
	assert.NoError(t, err)
	// prepare struct
	cfg := config.Config{}

	pb, err := pubsub.NewPubsub(1)
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

	// msg 8 update node label success
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
	err = h.OnMessage(msg8)
	assert.NoError(t, err)

	// msg 9 update node label err sub name
	msg9 := &specV1.Message{
		Kind: specV1.MessageCMD,
		Metadata: map[string]string{
			"cmd": specV1.MessageCommandNodeLabel,
		},
	}
	err = h.OnMessage(msg9)
	assert.Error(t, err, ErrSubNodeName)

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
