package rule

import (
	"fmt"
	"sync"

	"github.com/baidu/openedge/hub/common"
	"github.com/baidu/openedge/hub/router"
	"github.com/baidu/openedge/logger"
	"github.com/sirupsen/logrus"
)

type base interface {
	uid() string
	start() (err error)
	stop()
	wait(bool)
	channel() *msgchan
	register(sub *sinksub)
	remove(id, topic string)
	info() string
}

type rulebase struct {
	id      string
	sink    *sink
	broker  broker
	msgchan *msgchan
	once    sync.Once
	log     *logrus.Entry
}

func newRuleBase(id string, persistent bool, b broker, r *router.Trie, publish, republish common.Publish) *rulebase {
	log := logger.WithFields(common.LogComponent, "rule", common.LogTarget, id)
	rb := &rulebase{
		id:     id,
		broker: b,
		log:    log,
	}
	persist := rb.persist
	if !persistent {
		persist = nil
	}
	rb.msgchan = newMsgChan(
		b.Config().Message.Egress.Qos0.Buffer.Size,
		b.Config().Message.Egress.Qos1.Buffer.Size,
		publish,
		republish,
		b.Config().Message.Egress.Qos1.Retry.Interval,
		b.Config().Shutdown.Timeout,
		persist,
		log,
	)
	rb.sink = newSink(id, b, r, rb.msgchan)
	return rb
}

func newRuleQos0(b broker, r *router.Trie) *rulebase {
	return newRuleBase(common.RuleMsgQ0, false, b, r, nil, nil)
}

func newRuleTopic(b broker, r *router.Trie) *rulebase {
	rb := newRuleBase(common.RuleTopic, true, b, r, nil, nil)
	rb.msgchan.publish = rb.publish
	return rb
}

func newRuleSess(id string, p bool, b broker, r *router.Trie, publish, republish common.Publish) base {
	return newRuleBase(id, p, b, r, publish, republish)
}

func (r *rulebase) uid() string {
	return r.id
}

func (r *rulebase) publish(msg common.Message) {
	msg.QOS = msg.TargetQOS
	msg.Topic = msg.TargetTopic
	msg.SequenceID = 0
	if msg.QOS == 1 {
		msg.SetCallbackPID(0, func(_ uint32) { msg.Ack() })
	}
	r.broker.Flow(&msg)
}

func (r *rulebase) start() (err error) {
	r.once.Do(func() {
		err = r.msgchan.start()
		if err != nil {
			r.msgchan.close(true)
		}
		err = r.sink.start()
		if err != nil {
			r.stop()
			r.wait(true)
		}
	})
	return
}

func (r *rulebase) stop() {
	r.log.Debug("rule closing")
	r.sink.stop()
}

func (r *rulebase) wait(force bool) {
	r.sink.wait()
	r.msgchan.close(force)
	r.log.Debug("rule closed")
}

func (r *rulebase) channel() *msgchan {
	return r.msgchan
}

func (r *rulebase) register(sub *sinksub) {
	r.sink.register(sub)
}

func (r *rulebase) remove(id, topic string) {
	r.sink.remove(id, topic)
}

func (r *rulebase) persist(sid uint64) {
	err := r.broker.PersistOffset(r.id, sid)
	if err != nil {
		r.log.WithError(err).Errorf("failed to persist offset")
	}
}

func (r *rulebase) info() string {
	return fmt.Sprintf("offset:%d, msgq0:%d, msgq1:%d, msgack:%d, id: %s",
		r.sink.getOffset(), len(r.msgchan.msgq0), len(r.msgchan.msgq1), len(r.msgchan.msgack), r.uid())
}
