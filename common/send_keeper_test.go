// Package common 公共实现定义
package common

import (
	"fmt"
	"sync"
	"testing"
	"time"

	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestKeeper(t *testing.T) {
	var keeper SendKeeper
	ch := make(chan *specV1.Message, 50)
	send := func(msg *specV1.Message) error {
		ch <- msg
		return nil
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			msg := <-ch
			err := keeper.ReceiveResp(msg)
			assert.NoError(t, err)
		}
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			msg := &specV1.Message{
				Kind: specV1.MessageReport,
				Content: specV1.LazyValue{Value: specV1.Desire{
					"key": fmt.Sprint(i),
				}},
			}
			res, err := keeper.SendSync(msg, time.Second, send)
			assert.NoError(t, err)
			assert.EqualValues(t, msg, res)
		}
	}(&wg)
	wg.Wait()
}

func TestSendSync(t *testing.T) {
	var keeper SendKeeper
	send := func(msg *specV1.Message) error {
		reqID := msg.Metadata[RequestID]
		val, ok := keeper.results.Load(reqID)
		assert.True(t, ok)
		ch := val.(chan *specV1.Message)
		ch <- msg
		return nil
	}
	msg := &specV1.Message{
		Kind: specV1.MessageReport,
		Content: specV1.LazyValue{Value: specV1.Desire{
			"123": "456",
		}},
	}
	res, err := keeper.SendSync(msg, time.Second, send)
	assert.NoError(t, err)
	assert.EqualValues(t, msg, res)
}

func TestReceiveResp(t *testing.T) {
	var keeper SendKeeper
	msg := &specV1.Message{}
	err := keeper.ReceiveResp(msg)
	assert.Error(t, err)

	msg.Metadata = make(map[string]string)
	err = keeper.ReceiveResp(msg)
	assert.Error(t, err)

	msg.Metadata = make(map[string]string)
	msg.Metadata[RequestID] = uuid.New().String()
	err = keeper.ReceiveResp(msg)
	assert.Error(t, err)

	msg.Metadata = make(map[string]string)
	reqID := uuid.New().String()
	msg.Metadata[RequestID] = reqID
	keeper.results.LoadOrStore(reqID, make(chan *specV1.Message, 1))
	err = keeper.ReceiveResp(msg)
	assert.NoError(t, err)

	msg.Metadata = make(map[string]string)
	reqID = uuid.New().String()
	msg.Metadata[RequestID] = reqID
	msgCh := make(chan *specV1.Message, 1)
	keeper.results.LoadOrStore(reqID, msgCh)
	msgCh <- &specV1.Message{}
	err = keeper.ReceiveResp(msg)
	assert.NoError(t, err)
}
