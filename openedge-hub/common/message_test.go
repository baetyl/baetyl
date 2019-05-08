package common

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/baidu/openedge/openedge-hub/utils"
	"github.com/golang/protobuf/proto"
	"github.com/jpillora/backoff"
	"github.com/stretchr/testify/assert"
)

func TestMessageEncode(t *testing.T) {
	m := new(Message)
	assert.Equal(t, uint32(0), m.GetQOS())
	msg := NewMessage(1, "test", []byte("pld"), "client-1")
	msg.TargetQOS = 0
	msg.TargetTopic = "t1"
	msg.SequenceID = 2
	msg.PacketID = 10
	msg.Barrier = true

	persisted, err := proto.Marshal(&msg.Persisted)
	assert.NoError(t, err)
	msg1, err := UnmarshalMessage(utils.U64ToB(11), persisted)
	assert.NoError(t, err)

	assert.Equal(t, msg.GetQOS(), msg1.GetQOS())
	assert.Equal(t, msg.GetTopic(), msg1.GetTopic())
	assert.Equal(t, msg.GetPayload(), msg1.GetPayload())
	assert.Equal(t, msg.GetClientID(), msg1.GetClientID())

	assert.Equal(t, uint64(11), msg1.SequenceID)
	assert.Equal(t, uint32(0), msg1.TargetQOS)
	assert.Equal(t, "", msg1.TargetTopic)
	assert.Equal(t, uint32(0), msg1.PacketID)
	assert.Equal(t, false, msg1.Barrier)

	msg2 := &Transferred{Persisted: &msg.Persisted}
	assert.NoError(t, err)
	msg2.FunctionName = "filter"
	msg2.FunctionInvokeID = "uuid"
	msg2.FunctionInstanceID = "uuid-1"
	transferred, err := proto.Marshal(msg2)
	assert.NoError(t, err)
	msg3 := new(Transferred)
	err = proto.Unmarshal(transferred, msg3)
	assert.NoError(t, err)

	assert.Equal(t, msg2.GetPersisted().GetQOS(), msg3.GetPersisted().GetQOS())
	assert.Equal(t, msg2.GetPersisted().GetTopic(), msg3.GetPersisted().GetTopic())
	assert.Equal(t, msg2.GetPersisted().GetPayload(), msg3.GetPersisted().GetPayload())
	assert.Equal(t, msg2.GetPersisted().GetClientID(), msg3.GetPersisted().GetClientID())

	assert.Equal(t, "filter", msg3.GetFunctionName())
	assert.Equal(t, "uuid", msg3.GetFunctionInvokeID())
	assert.Equal(t, "uuid-1", msg3.GetFunctionInstanceID())
}

func TestMessageAckWait(t *testing.T) {
	last := uint64(0)
	count := uint64(0)
	callback := func(sid uint64) { atomic.StoreUint64(&last, sid) }
	redo := func(Message) { atomic.AddUint64(&count, 1) }
	acks := [3]*MsgAck{
		&MsgAck{Message: &Message{SequenceID: 1}, FST: time.Now()},
		&MsgAck{Message: &Message{SequenceID: 2}, FST: time.Now()},
		&MsgAck{Message: &Message{SequenceID: 3}, FST: time.Now()},
	}
	acks[0].SetCallbackSID(callback)
	acks[1].SetCallbackSID(callback)
	acks[2].SetCallbackSID(callback)
	acks[0].SetAcknowledge()
	acks[1].SetAcknowledge()
	acks[2].SetAcknowledge()
	// expected := []bool{false, true, false}
	c := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bf := &backoff.Backoff{
			Min:    time.Millisecond * 100,
			Max:    time.Millisecond * 100,
			Factor: 2,
		}
		for _, ack := range acks {
			ack.WaitTimeout(bf, redo, c)
		}
	}()
	assert.Equal(t, uint64(0), atomic.LoadUint64(&last))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&count))
	acks[0].Ack()
	time.Sleep(time.Millisecond * 150)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&last))
	assert.Equal(t, uint64(1), atomic.LoadUint64(&count))
	acks[1].Ack()
	acks[2].Ack()
	wg.Wait()
	assert.Equal(t, uint64(3), atomic.LoadUint64(&last))
	assert.Equal(t, uint64(1), atomic.LoadUint64(&count))
}

func TestMessageAckRedoClose(t *testing.T) {
	actual := uint64(0)
	c := make(chan struct{})
	ack := &MsgAck{Message: &Message{SequenceID: 10}, FST: time.Now()}
	ack.SetCallbackSID(func(sid uint64) { atomic.StoreUint64(&actual, sid) })
	ack.SetAcknowledge()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bf := &backoff.Backoff{
			Min:    time.Millisecond * 100,
			Max:    time.Millisecond * 100,
			Factor: 2,
		}
		ack.WaitTimeout(bf, func(_ Message) { atomic.AddUint64(&actual, 1) }, c)
	}()
	assert.Equal(t, uint64(0), atomic.LoadUint64(&actual))
	time.Sleep(time.Millisecond * 150)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&actual))
	time.Sleep(time.Millisecond * 100)
	close(c)
	wg.Wait()
	assert.Equal(t, uint64(2), atomic.LoadUint64(&actual))
}
