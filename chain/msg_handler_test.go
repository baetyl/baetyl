package chain

import (
	"io"
	"sync"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pubsub"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/ami"
)

const (
	ErrUnmarshal = "failed to unmarshal data message"
	ErrWrite     = "failed to write debug command"
	ErrTimeout   = "chain timeout"
)

var (
	handlerWG = sync.WaitGroup{}
)

func TestHandler(t *testing.T) {
	pb, err := pubsub.NewPubsub(1)
	assert.NoError(t, err)

	pipe := ami.Pipe{}
	pipe.InReader, pipe.InWriter = io.Pipe()
	pipe.OutReader, pipe.OutWriter = io.Pipe()

	var opt ami.DebugOptions
	opt.KubeDebugOptions = ami.KubeDebugOptions{
		Namespace: "default",
		Name:      "baetyl-function-0",
		Container: "function",
		Command:   []string{},
	}
	cha := &chain{
		debugOptions: &opt,
		token:        token,
		upside:       "up",
		pb:           pb,
		pipe:         pipe,
		log:          log.L().With(log.Any("chain", "test")),
	}

	cHandler := &chainHandler{chain: cha}

	ch, err := cha.pb.Subscribe(cha.upside)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	pro := pubsub.NewProcessor(ch, 0, &msgUpside{t: t, pb: cha.pb})
	pro.Start()

	msg0 := &specV1.Message{
		Kind: specV1.MessageCMD,
	}
	err = cHandler.OnMessage(msg0)
	assert.NoError(t, err)

	handlerWG.Add(1)
	msg1 := &specV1.Message{
		Kind:    specV1.MessageData,
		Content: specV1.LazyValue{Value: 1},
	}
	err = cHandler.OnMessage(msg1)
	assert.Error(t, err)

	err = cha.pipe.InWriter.Close()
	assert.NoError(t, err)
	handlerWG.Add(1)
	msg2 := &specV1.Message{
		Kind:    specV1.MessageData,
		Content: specV1.LazyValue{Value: testMsg},
	}
	err = cHandler.OnMessage(msg2)
	assert.Error(t, err)

	handlerWG.Add(1)

	msg3 := &specV1.Message{
		Kind:    specV1.MessageData,
		Content: specV1.LazyValue{Value: []byte(ExitCmd)},
	}
	err = cHandler.OnMessage(msg3)
	assert.NoError(t, err)

	err = cHandler.OnTimeout()
	assert.NoError(t, err)

	handlerWG.Wait()
	pro.Close()
}

type msgUpside struct {
	t  *testing.T
	pb pubsub.Pubsub
}

func (h *msgUpside) OnMessage(msg interface{}) error {
	m, ok := msg.(*specV1.Message)
	assert.True(h.t, ok)
	assert.Equal(h.t, "false", m.Metadata["success"])

	switch m.Metadata["msg"] {
	case ErrWrite, ErrUnmarshal, ErrTimeout:
		handlerWG.Done()
	default:
		assert.Fail(h.t, "unexpected message")
	}
	return nil
}

func (h *msgUpside) OnTimeout() error {
	return nil
}
