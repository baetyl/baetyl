package rule

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	bb "github.com/baidu/openedge/hub/broker"
	"github.com/baidu/openedge/hub/common"
	"github.com/baidu/openedge/hub/config"
	"github.com/baidu/openedge/hub/persist"
	"github.com/baidu/openedge/hub/utils"
	"github.com/baidu/openedge/logger"
	"github.com/stretchr/testify/assert"
)

func TestBroker(t *testing.T) {
	os.RemoveAll("./var/db")

	c, _ := config.NewConfig([]byte(""))
	// c.Logger.Console = true
	// c.Logger.Level = "debug"
	assert.NoError(t, logger.Init(c.Logger))
	pf, err := persist.NewFactory("./var/db/")
	assert.NoError(t, err)
	defer pf.Close()
	b, err := bb.NewBroker(c, pf)
	assert.NoError(t, err)
	defer b.Close()
	subs := make([]config.Subscription, 2)
	subs[0].Source.Topic = "head"
	subs[0].Target.Topic = "next"
	subs[1].Source.Topic = "next"
	subs[1].Target.Topic = "tail"
	r, err := NewManager(subs, b)
	assert.NoError(t, err)
	defer r.Close()
	r.Start()
	q0 := make(chan *common.Message, 10)
	q1 := make(chan *common.Message, 10)
	pidchan := make(chan uint32, 10)
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
	err = r.AddSinkSub("$session/tail", "$session/tail", 0, "tail", 0, "")
	assert.NoError(t, err)
	err = r.StartRule("$session/tail")
	assert.NoError(t, err)

	fmt.Println("--> start")
	msg := common.NewMessage(0, "head", []byte("pld:3"), "cid")
	b.Flow(msg)
	out := <-q0
	assert.Equal(t, int(0), int(out.SequenceID))
	assert.Equal(t, "tail", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, "pld:3", string(out.Payload))
	msgcb := common.NewMessage(1, "head", []byte("pld:4"), "cid")
	msgcb.SetCallbackPID(2, func(pid uint32) { pidchan <- pid })
	b.Flow(msgcb)
	assert.Equal(t, uint32(2), <-pidchan)
	out = <-q0
	assert.Equal(t, int(0), int(out.SequenceID))
	assert.Equal(t, "tail", out.Topic)
	assert.Equal(t, uint32(0), out.QOS)
	assert.Equal(t, "pld:4", string(out.Payload))

	select {
	case <-q0:
		assert.Fail(t, "message not expected")
	case <-q1:
		assert.Fail(t, "message not expected")
	case <-time.After(time.Second):
	}

	offset, err := b.OffsetPersisted(common.RuleTopic)
	assert.NoError(t, err)
	assert.NotNil(t, offset)
	assert.Equal(t, int(1), int(*offset))
	offset, err = b.OffsetPersisted("$session/tail")
	assert.NoError(t, err)
	assert.Nil(t, offset)
	offset, err = b.OffsetPersisted("nonexist")
	assert.NoError(t, err)
	assert.Nil(t, offset)

	fmt.Println("--> Remove rule ($session/tail)")
	err = r.RemoveRule("$session/tail")
	assert.NoError(t, err)
	msg = common.NewMessage(0, "head", []byte("pld:7"), "cid")
	b.Flow(msg)
	msgcb = common.NewMessage(1, "head", []byte("pld:8"), "cid")
	msgcb.SetCallbackPID(3, func(pid uint32) { pidchan <- pid })
	b.Flow(msgcb)
	select {
	case <-q0:
		assert.Fail(t, "message not expected")
	case <-q1:
		assert.Fail(t, "message not expected")
	case <-time.After(time.Second):
	}
	offset, err = b.OffsetPersisted(common.RuleTopic)
	assert.NoError(t, err)
	assert.NotNil(t, offset)
	assert.Equal(t, int(2), int(*offset))
	offset, err = b.OffsetPersisted("$session/tail")
	assert.NoError(t, err)
	assert.Nil(t, offset)
	offset, err = b.OffsetPersisted("nonexist")
	assert.NoError(t, err)
	assert.Nil(t, offset)

	fmt.Println("--> Remove rule topic")
	err = r.AddRuleSess("$session/tail", true,
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
	err = r.AddSinkSub("$session/tail", "$session/tail", 0, "tail", 0, "")
	assert.NoError(t, err)
	err = r.StartRule("$session/tail")
	assert.NoError(t, err)
	err = r.RemoveRule(common.RuleTopic)
	assert.NoError(t, err)
	msg = common.NewMessage(0, "head", []byte("pld:7"), "cid")
	b.Flow(msg)
	msgcb = common.NewMessage(1, "head", []byte("pld:8"), "cid")
	msgcb.SetCallbackPID(3, func(pid uint32) { pidchan <- pid })
	b.Flow(msgcb)
	select {
	case <-q0:
		assert.Fail(t, "message not expected")
	case <-q1:
		assert.Fail(t, "message not expected")
	case <-time.After(time.Second):
	}
	offset, err = b.OffsetPersisted(common.RuleTopic)
	assert.NoError(t, err)
	assert.NotNil(t, offset)
	assert.Equal(t, int(2), int(*offset))
	offset, err = b.OffsetPersisted("$session/tail")
	assert.NoError(t, err)
	assert.NotNil(t, offset)
	assert.Equal(t, int(3), int(*offset))
	offset, err = b.OffsetPersisted("nonexist")
	assert.NoError(t, err)
	assert.Nil(t, offset)
}

//func TestBrokerClose(t *testing.T) {
//	os.RemoveAll("./var/db")
//
//	c, _ := config.NewConfig([]byte(""))
//	// c.Logger.Console = true
//	// c.Logger.Level = "debug"
//	assert.NoError(t, logger.Init(c.Logger))
//	pf, err := persist.NewFactory("./var/db/")
//	assert.NoError(t, err)
//	defer pf.Close()
//	b, err := bb.NewBroker(c, pf)
//	assert.NoError(t, err)
//	subs := make([]config.Subscription, 0)
//	r, err := NewManager(subs, b)
//	assert.NoError(t, err)
//	q0 := make(chan *common.Message, 10)
//	q1 := make(chan *common.Message, 10)
//	err = r.AddRuleSess("$session", true,
//		func(msg common.Message) {
//			if msg.QOS == 0 {
//				q0 <- &msg
//			} else {
//				q1 <- &msg
//			}
//			msg.Ack()
//		},
//		nil)
//	assert.NoError(t, err)
//	err = r.AddSinkSub("$session", "$session", 1, "in", 1, "")
//	assert.NoError(t, err)
//	r.Start()
//
//	fmt.Println("--> starting")
//	msgin := common.NewMessage(0, "in", []byte("01"), "cid0")
//	msgackin := common.NewMessage(1, "in", []byte("11"), "cid1")
//	msgackin.SetCallbackPID(1, nil)
//	b.Flow(msgin)
//	b.Flow(msgackin)
//	out0 := <-q0
//	assert.Equal(t, uint32(0), out0.QOS)
//	assert.Equal(t, uint64(0), out0.SequenceID)
//	assert.Equal(t, "cid0", out0.ClientID)
//	out1 := <-q1
//	assert.Equal(t, uint32(1), out1.QOS)
//	assert.Equal(t, uint64(1), out1.SequenceID)
//	assert.Equal(t, "cid1", out1.ClientID)
//	select {
//	case <-q0:
//		assert.Fail(t, "Message unexpected")
//	case <-q0:
//		assert.Fail(t, "Message unexpected")
//	case <-time.After(time.Millisecond * 10):
//	}
//
//	fmt.Println("--> close rule manager")
//	r.Close()
//	b.Flow(msgin)
//	b.Flow(msgackin)
//
//	fmt.Println("--> start rule manager")
//	r, err = NewManager(subs, b)
//	assert.NoError(t, err)
//	err = r.AddRuleSess("$session", true,
//		func(msg common.Message) {
//			if msg.QOS == 0 {
//				q0 <- &msg
//			} else {
//				q1 <- &msg
//			}
//			msg.Ack()
//		},
//		nil)
//	assert.NoError(t, err)
//	err = r.AddSinkSub("$session", "$session", 1, "in", 1, "")
//	assert.NoError(t, err)
//	r.Start()
//	msgin = common.NewMessage(0, "in", []byte("03"), "cid00")
//	msgackin = common.NewMessage(1, "in", []byte("13"), "cid11")
//	b.Flow(msgin)
//	b.Flow(msgackin)
//	out0 = <-q0
//	assert.Equal(t, uint32(0), out0.QOS)
//	assert.Equal(t, uint64(0), out0.SequenceID)
//	assert.Equal(t, "cid0", out0.ClientID)
//	out0 = <-q0
//	assert.Equal(t, uint32(0), out0.QOS)
//	assert.Equal(t, uint64(0), out0.SequenceID)
//	assert.Equal(t, "cid00", out0.ClientID)
//	out1 = <-q1
//	assert.Equal(t, uint32(1), out1.QOS)
//	assert.Equal(t, uint64(2), out1.SequenceID)
//	assert.Equal(t, "cid1", out1.ClientID)
//	out1 = <-q1
//	assert.Equal(t, uint32(1), out1.QOS)
//	assert.Equal(t, uint64(3), out1.SequenceID)
//	assert.Equal(t, "cid11", out1.ClientID)
//	select {
//	case <-q0:
//		assert.Fail(t, "Message unexpected")
//	case <-q1:
//		assert.Fail(t, "Message unexpected")
//	case <-time.After(time.Millisecond * 10):
//	}
//
//	fmt.Println("--> close broker")
//	r.Close()
//	b.Close()
//
//	fmt.Println("--> start broker")
//	b, err = bb.NewBroker(c, pf)
//	assert.NoError(t, err)
//	r, err = NewManager(subs, b)
//	assert.NoError(t, err)
//	err = r.AddRuleSess("$session", true,
//		func(msg common.Message) {
//			if msg.QOS == 0 {
//				q0 <- &msg
//			} else {
//				q1 <- &msg
//			}
//			msg.Ack()
//		},
//		nil)
//	assert.NoError(t, err)
//	err = r.AddSinkSub("$session", "$session", 1, "in", 1, "")
//	assert.NoError(t, err)
//	r.Start()
//	msgin = common.NewMessage(0, "in", []byte("04"), "cid000")
//	msgackin = common.NewMessage(1, "in", []byte("14"), "cid111")
//	msgackin.SetCallbackPID(1, nil)
//	b.Flow(msgin)
//	b.Flow(msgackin)
//	out0 = <-q0
//	assert.Equal(t, uint32(0), out0.QOS)
//	assert.Equal(t, uint64(0), out0.SequenceID)
//	assert.Equal(t, "cid000", out0.ClientID)
//	out1 = <-q1
//	assert.Equal(t, uint32(1), out1.QOS)
//	assert.Equal(t, uint64(4), out1.SequenceID)
//	assert.Equal(t, "cid111", out1.ClientID)
//	select {
//	case <-q0:
//		assert.Fail(t, "Message unexpected")
//	case <-q1:
//		assert.Fail(t, "Message unexpected")
//	case <-time.After(time.Millisecond * 10):
//	}
//	r.Close()
//	b.Close()
//}

func TestBrokerCleaning(t *testing.T) {
	os.RemoveAll("./var/db")

	c, _ := config.NewConfig([]byte(""))
	// c.Logger.Console = true
	// c.Logger.Level = "debug"
	c.Message.Ingress.Qos1.Cleanup.Interval = time.Second
	c.Message.Ingress.Qos1.Cleanup.Retention = time.Second
	assert.NoError(t, logger.Init(c.Logger))
	pf, err := persist.NewFactory("./var/db/")
	assert.NoError(t, err)
	defer pf.Close()
	b, err := bb.NewBroker(c, pf)
	assert.NoError(t, err)
	defer b.Close()
	subs := make([]config.Subscription, 0)
	r, err := NewManager(subs, b)
	assert.NoError(t, err)
	defer r.Close()
	q0 := make(chan *common.Message, 10)
	q1 := make(chan *common.Message, 10)
	err = r.AddRuleSess("$session", true,
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
	err = r.AddSinkSub("$session", "$session", 1, "in", 1, "")
	assert.NoError(t, err)
	r.Start()

	fmt.Println("--> starting")
	msgackin := common.NewMessage(1, "in", []byte("11"), "cid")
	b.Flow(msgackin)
	msg := <-q1
	assert.Equal(t, int(1), int(msg.SequenceID))
	time.Sleep(time.Second * 2)
	msgackin = common.NewMessage(1, "in", []byte("12"), "cid")
	msgackin.SetCallbackPID(1, nil)
	b.Flow(msgackin)
	msg = <-q1
	assert.Equal(t, int(2), int(msg.SequenceID))
	db, err := pf.NewDB("msgqos1.db")
	assert.NoError(t, err)
	kvs, err := db.BatchFetch(utils.U64ToB(uint64(0)), 1)
	assert.NoError(t, err)
	assert.Equal(t, int(2), int(utils.U64(kvs[0].Key)))
}

func TestBrokerPerf(t *testing.T) {
	t.Skip("Skip perf test")
	os.RemoveAll("./var/db")
	c, _ := config.NewConfig([]byte(""))
	c.Logger.Console = true
	// c.Logger.Level = "debug"
	assert.NoError(t, logger.Init(c.Logger))
	pf, err := persist.NewFactory("./var/db/")
	assert.NoError(t, err)
	defer pf.Close()
	b, err := bb.NewBroker(c, pf)
	assert.NoError(t, err)
	defer b.Close()
	subs := make([]config.Subscription, 2)
	subs[0].Source.QOS = 1
	subs[0].Source.Topic = "head"
	subs[0].Target.QOS = 1
	subs[0].Target.Topic = "tail"
	r, err := NewManager(subs, b)
	assert.NoError(t, err)
	defer r.Close()
	r.Start()

	total := 500000
	q0 := make(chan *common.Message, total)
	q1 := make(chan *common.Message, total)
	err = r.AddRuleSess("$session/tail", false,
		func(msg common.Message) {
			if msg.QOS == 0 {
				q0 <- &msg
			} else {
				q1 <- &msg
			}
			msg.Ack()
		}, nil)
	assert.NoError(t, err)
	err = r.AddSinkSub("$session/tail", "$session/tail", 1, "tail", 1, "")
	assert.NoError(t, err)
	err = r.StartRule("$session/tail")
	assert.NoError(t, err)

	exit := make(chan struct{}, 0)
	pubFinished := int32(0)
	subCount := int32(0)
	go func() {
		defer atomic.StoreInt32(&pubFinished, 1)
		msg := common.NewMessage(1, "head", []byte("pld:1"), "cid")
		for i := 0; i < total; i++ {
			b.Flow(msg)
		}
		fmt.Println("Pub finished")
	}()

	go func() {
		for i := 0; i < total; i++ {
			select {
			case <-exit:
				return
			case <-q1:
				atomic.AddInt32(&subCount, 1)
			}
		}
	}()

	var pT, cT time.Time
	var pC, cC int32
	for {
		select {
		case <-time.After(time.Second):
			if pT.IsZero() {
				pT = time.Now()
				pC = atomic.LoadInt32(&subCount)
				continue
			}
			cT = time.Now()
			cC = atomic.LoadInt32(&subCount)
			mps := float64(cC-pC) / cT.Sub(pT).Seconds()
			pT = cT
			pC = cC
			fmt.Printf("Sub MPS: %f\n", mps)
			if mps < 90 {
				return
			}
		}
	}
}
