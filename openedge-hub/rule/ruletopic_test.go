package rule

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bb "github.com/baidu/openedge/openedge-hub/broker"
	"github.com/baidu/openedge/openedge-hub/common"
	"github.com/baidu/openedge/openedge-hub/config"
	"github.com/baidu/openedge/openedge-hub/persist"
	"github.com/jpillora/backoff"
	"github.com/stretchr/testify/assert"
)

func TestRuleTopicFanout(t *testing.T) {
	os.RemoveAll("./var/db")
	c, _ := config.NewConfig([]byte(""))
	// c.Logger.Console = true
	// c.Logger.Level = "debug"
	// assert.NoError(t, logger.Init(c.Logger))
	pf, err := persist.NewFactory("./var/db/")
	assert.NoError(t, err)
	defer pf.Close()
	b, err := bb.NewBroker(c, pf)
	assert.NoError(t, err)
	defer b.Close()
	subs := make([]config.Subscription, 2)
	subs[0].Source.Topic = "all"
	subs[0].Target.Topic = "00"
	subs[1].Source.QOS = 1
	subs[1].Source.Topic = "all"
	subs[1].Target.QOS = 1
	subs[1].Target.Topic = "11"
	r, err := NewManager(subs, b)
	assert.NoError(t, err)
	defer r.Close()
	r.Start()
	q0 := make(chan *common.Message, 10)
	q1 := make(chan *common.Message, 10)
	err = r.AddRuleSess("$session/tail", false,
		func(msg common.Message) {
			if msg.QOS == 0 {
				q0 <- &msg
			} else {
				q1 <- &msg
			}
			msg.Ack()
		},
		nil)
	assert.NoError(t, err)
	err = r.AddSinkSub("$session/tail", "$session/tail", 0, "#", 0, "")
	assert.NoError(t, err)
	err = r.StartRule("$session/tail")
	assert.NoError(t, err)

	fmt.Println("--> start")
	b.Flow(common.NewMessage(0, "all", []byte("pld"), ""))
	out := <-q0
	assert.Equal(t, "all", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, int(0), int(out.SequenceID))
	out0 := <-q0
	assert.Equal(t, uint32(0), out0.QOS)
	assert.Equal(t, int(0), int(out0.SequenceID))
	out1 := <-q0
	assert.Equal(t, uint32(0), out1.QOS)
	assert.Equal(t, int(0), int(out1.SequenceID))
	if out0.Topic == "00" {
		assert.Equal(t, "11", out1.Topic)
	} else if out0.Topic == "11" {
		assert.Equal(t, "00", out1.Topic)
	} else {
		assert.FailNow(t, "Not expected")
	}
	b.Flow(common.NewMessage(1, "all", []byte("pld"), ""))
	out = <-q1
	assert.Equal(t, "all", out.Topic)
	assert.Equal(t, uint32(1), out.QOS)
	assert.Equal(t, int(1), int(out.SequenceID))
	out = <-q0
	assert.Equal(t, "00", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, int(0), int(out.SequenceID))
	out = <-q1
	assert.Equal(t, "11", out.Topic)
	assert.Equal(t, uint32(1), out.QOS)
	assert.Equal(t, int(2), int(out.SequenceID))
	select {
	case <-q0:
		assert.Fail(t, "message not expected")
	case <-q1:
		assert.Fail(t, "message not expected")
	case <-time.After(time.Second):
	}
}

func TestRuleTopicFanin(t *testing.T) {
	os.RemoveAll("./var/db")
	c, _ := config.NewConfig([]byte(""))
	// c.Logger.Console = true
	// c.Logger.Level = "debug"
	// assert.NoError(t, logger.Init(c.Logger))
	pf, err := persist.NewFactory("./var/db/")
	assert.NoError(t, err)
	defer pf.Close()
	b, err := bb.NewBroker(c, pf)
	assert.NoError(t, err)
	defer b.Close()
	subs := make([]config.Subscription, 2)
	subs[0].Source.Topic = "00"
	subs[0].Target.Topic = "all"
	subs[1].Source.QOS = 1
	subs[1].Source.Topic = "11"
	subs[1].Target.QOS = 1
	subs[1].Target.Topic = "all"
	r, err := NewManager(subs, b)
	assert.NoError(t, err)
	defer r.Close()
	r.Start()
	q0 := make(chan *common.Message, 10)
	q1 := make(chan *common.Message, 10)
	err = r.AddRuleSess("$session/tail", false,
		func(msg common.Message) {
			if msg.QOS == 0 {
				q0 <- &msg
			} else {
				q1 <- &msg
			}
			msg.Ack()
		},
		nil)
	assert.NoError(t, err)
	err = r.AddSinkSub("$session/tail", "$session/tail", 1, "#", 1, "")
	assert.NoError(t, err)
	err = r.StartRule("$session/tail")
	assert.NoError(t, err)

	fmt.Println("--> start")
	b.Flow(common.NewMessage(0, "00", []byte("pld"), ""))
	b.Flow(common.NewMessage(0, "11", []byte("pld"), ""))
	out := <-q0
	assert.Equal(t, "00", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, int(0), int(out.SequenceID))
	out0 := <-q0
	assert.Equal(t, uint32(0), out0.QOS)
	assert.Equal(t, int(0), int(out0.SequenceID))
	out1 := <-q0
	assert.Equal(t, uint32(0), out1.QOS)
	assert.Equal(t, int(0), int(out1.SequenceID))
	if out0.Topic == "11" {
		assert.Equal(t, "all", out1.Topic)
	} else if out0.Topic == "all" {
		assert.Equal(t, "11", out1.Topic)
	} else {
		assert.FailNow(t, "Not expected")
	}
	out = <-q0
	assert.Equal(t, "all", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, int(0), int(out.SequenceID))

	b.Flow(common.NewMessage(1, "00", []byte("pld"), ""))
	b.Flow(common.NewMessage(1, "11", []byte("pld"), ""))
	out = <-q1
	assert.Equal(t, "00", out.Topic)
	assert.Equal(t, uint32(1), out.QOS)
	assert.Equal(t, int(1), int(out.SequenceID))
	out = <-q1
	assert.Equal(t, "11", out.Topic)
	assert.Equal(t, uint32(1), out.QOS)
	assert.Equal(t, int(2), int(out.SequenceID))
	out = <-q0
	assert.Equal(t, "all", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, int(0), int(out.SequenceID))
	out = <-q1
	assert.Equal(t, "all", out.Topic)
	assert.Equal(t, uint32(1), out.QOS)
	assert.Equal(t, int(3), int(out.SequenceID))
	select {
	case <-q0:
		assert.Fail(t, "message not expected")
	case <-q1:
		assert.Fail(t, "message not expected")
	case <-time.After(time.Second):
	}
}

func TestRuleTopicRedo(t *testing.T) {
	os.RemoveAll("./var/db")
	c, _ := config.NewConfig([]byte(""))
	// c.Logger.Console = true
	// c.Logger.Level = "debug"
	// assert.NoError(t, logger.Init(c.Logger))
	pf, err := persist.NewFactory("./var/db/")
	assert.NoError(t, err)
	defer pf.Close()
	b, err := bb.NewBroker(c, pf)
	assert.NoError(t, err)
	defer b.Close()
	subs := make([]config.Subscription, 2)
	subs[0].Source.QOS = 1
	subs[0].Source.Topic = "a"
	subs[0].Target.QOS = 1
	subs[0].Target.Topic = "b"
	r, err := NewManager(subs, b)
	assert.NoError(t, err)
	defer r.Close()
	tmp, ok := r.rules.Get(common.RuleTopic)
	assert.True(t, ok)
	rt, ok := tmp.(*rulebase)
	assert.True(t, ok)
	rc := int32(0)
	var wg sync.WaitGroup
	wg.Add(1)
	rt.msgchan.republishBackoff = &backoff.Backoff{
		Min:    time.Millisecond * 100,
		Max:    time.Millisecond * 100,
		Factor: 2,
	}
	rt.msgchan.publish = func(msg common.Message) {}
	rt.msgchan.republish = func(msg common.Message) {
		atomic.AddInt32(&rc, 1)
		assert.Equal(t, "a", msg.Topic)
		assert.Equal(t, uint32(1), msg.QOS)
		assert.Equal(t, "b", msg.TargetTopic)
		assert.Equal(t, uint32(1), msg.QOS)
		wg.Done()
	}
	r.Start()
	fmt.Println("--> start")
	b.Flow(common.NewMessage(1, "a", []byte("pld"), ""))
	wg.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&rc))
}
