package mqttlink

import (
	"context"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/common"
	"github.com/baetyl/baetyl/v2/plugin"
)

func TestMqttLink_State(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	link := &mqttLink{
		keeper: common.SendKeeper{},
		msgCh:  make(chan *specV1.Message, 1),
		errCh:  make(chan error, 1),
		ctx:    ctx,
		state:  &specV1.Message{Kind: plugin.LinkStateUnknown, Content: specV1.LazyValue{Value: ""}},
		obsCh:  make(chan *specV1.Message, 1024),
		cancel: cancel,
		log:    log.With(log.Any("plugin", "mqttlink")),
	}

	// case LinkStateUnknown
	res := link.State()
	assert.Equal(t, plugin.LinkStateUnknown, string(res.Kind))

	// case LinkStateNetworkError
	err := link.dial()
	assert.Error(t, err)
	res = link.State()
	assert.Equal(t, plugin.LinkStateNetworkError, string(res.Kind))

	go link.receiving()
	go func() {
		for {
			select {
			case <-link.errCh:
			case <-link.ctx.Done():
				return
			}
		}
	}()

	// case LinkStateNodeNotFound
	nodeNotFoundErrMsg := &specV1.Message{
		Kind: specV1.MessageError,
		Metadata: map[string]string{
			"name": "test",
		},
		Content: specV1.LazyValue{Value: "The (node) resource (test) is not found."},
	}
	dt, err := nodeNotFoundErrMsg.Content.MarshalJSON()
	assert.NoError(t, err)
	nodeNotFoundErrMsg.Content.SetJSON(dt)
	link.obsCh <- nodeNotFoundErrMsg
	time.Sleep(time.Second)
	res = link.State()
	assert.Equal(t, plugin.LinkStateNodeNotFound, string(res.Kind))

	// case LinkStateSucceeded
	nodeOtherErrMsg := &specV1.Message{
		Kind: specV1.MessageError,
		Metadata: map[string]string{
			"name": "test",
		},
		Content: specV1.LazyValue{Value: "xxx"},
	}
	dt, err = nodeOtherErrMsg.Content.MarshalJSON()
	assert.NoError(t, err)
	nodeOtherErrMsg.Content.SetJSON(dt)
	link.obsCh <- nodeOtherErrMsg
	time.Sleep(time.Second)
	res = link.State()
	assert.Equal(t, plugin.LinkStateSucceeded, string(res.Kind))

	link.ctx.Done()
}
