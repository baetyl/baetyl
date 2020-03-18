package hub

import (
	"fmt"
	"time"

	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/baetyl-hub/utils"
	"github.com/baetyl/baetyl/logger"
	"github.com/jpillora/backoff"
)

var errMsgChanClosed = fmt.Errorf("message channel is closed")
var errMsgDiscarded = fmt.Errorf("message is discarded since channel is full")

// Q0: it means the message with QOS=0 published by clients or functions
// Q1: it means the message with QOS=1 published by clients or functions

// msgchan message channel routed from sink
type msgchan struct {
	msgq0            chan *common.Message
	msgq1            chan *common.Message
	msgack           chan *common.MsgAck // TODO: move to sink?
	persist          func(uint64)
	msgtomb          utils.Tomb
	acktomb          utils.Tomb
	quitTimeout      time.Duration
	publish          common.Publish
	republish        common.Publish
	republishBackoff *backoff.Backoff
	log              logger.Logger
}

// newMsgChan creates a new message channel
func newMsgChan(l0, l1 int, publish, republish common.Publish, republishTimeout time.Duration, quitTimeout time.Duration, persist func(uint64), log logger.Logger) *msgchan {
	backoff := &backoff.Backoff{
		Min:    time.Millisecond * 100,
		Max:    republishTimeout,
		Factor: 2,
	}
	return &msgchan{
		msgq0:            make(chan *common.Message, l0),
		msgq1:            make(chan *common.Message, l1),
		msgack:           make(chan *common.MsgAck, l1),
		publish:          publish,
		republish:        republish,
		republishBackoff: backoff,
		quitTimeout:      quitTimeout,
		persist:          persist,
		log:              log,
	}
}

// MsgQ0ChanLen returns length of channel of message published with qos=0
func (c *msgchan) msgq0ChanLen() int {
	return len(c.msgq0)
}

// MsgQ1ChanLen returns length of channel of message published with qos=1
func (c *msgchan) msgq1ChanLen() int {
	return len(c.msgq1)
}

func (c *msgchan) start() error {
	err := c.acktomb.Gos(c.goWaitingAck)
	if err != nil {
		return err
	}
	return c.msgtomb.Gos(c.goProcessingQ0, c.goProcessingQ1)
}

// TODO: only handle messages with qos 0 before system closed. [bce-iot-6545]
func (c *msgchan) close(force bool) {
	c.log.Debugf("message channel closing")

	c.msgtomb.Kill()
	if force {
		c.acktomb.Kill()
	}
	err := c.msgtomb.Wait()
	if err != nil {
		c.log.WithError(err).Debugf("message channel closed")
	}
	if !force {
		c.acktomb.Kill()
	}
	err = c.acktomb.Wait()
	if err != nil {
		c.log.WithError(err).Debugf("message channel closed")
	}
}

// PutQ0 put message published with qos=0
func (c *msgchan) putQ0(msg *common.Message) {
	if !c.msgtomb.Alive() {
		c.log.WithError(errMsgChanClosed).Errorf("failed to put message (qos=0)")
		return
	}
	select {
	case c.msgq0 <- msg:
	default: // discard if channel is full
		c.discard(msg)
	}
}

// PutQ1 put message published with qos=1
func (c *msgchan) putQ1(msg *common.Message) {
	select {
	case <-c.msgtomb.Dying():
		c.log.WithError(errMsgChanClosed).Errorf("failed to put message (qos=1)")
	case c.msgq1 <- msg:
	}
}

// ProcessingQ0 processing message with QOS=0
func (c *msgchan) goProcessingQ0() error {
	c.log.Debugf("task of processing message from channel (Q0) begins")
	defer c.log.Debugf("task of processing message from channel (Q0) stopped")

loop:
	for {
		select {
		case <-c.msgtomb.Dying():
			break loop
		case msg := <-c.msgq0:
			c.process(msg)
		}
	}
	// Try to handle all message with qos=0
	endTime := time.Now().Add(c.quitTimeout)
	for {
		select {
		case msg := <-c.msgq0:
			c.process(msg)
		case <-time.After(endTime.Sub(time.Now())):
			c.log.Warnf("timed out to process inflight messages from channel (Q0) during shutdown")
			return nil
		case <-c.acktomb.Dying():
			c.log.Debugf("interrupted to process inflight messages from channel (Q0) during session close")
			return nil
		default:
			c.log.Debugf("finished to process inflight messages from channel (Q0)")
			return nil
		}
	}
}

// ProcessingQ1 processing message with QOS=1
func (c *msgchan) goProcessingQ1() error {
	c.log.Debugf("task of processing message from channel (Q1) begins")
	defer c.log.Debugf("task of processing message from channel (Q1) stopped")

loop:
	for {
		select {
		case <-c.msgtomb.Dying():
			break loop
		case msg := <-c.msgq1:
			c.process(msg)
		}
	}
	// Try to handle all message with qos=1
	endTime := time.Now().Add(c.quitTimeout)
	for {
		select {
		case msg := <-c.msgq1:
			c.process(msg)
		case <-time.After(endTime.Sub(time.Now())):
			c.log.Warnf("timed out to process inflight messages (qos=1) during shutdown")
			return nil
		case <-c.acktomb.Dying():
			c.log.Debugf("interrupted to process inflight messages (qos=1) during session close")
			return nil
		default:
			c.log.Debugf("finished to process inflight messages (qos=1)")
			return nil
		}
	}
}

func (c *msgchan) goWaitingAck() error {
	c.log.Debugf("task of waiting acknowledge begins")
	defer c.log.Debugf("task of waiting acknowledge stopped")

	for {
		select {
		case <-c.acktomb.Dying():
			return nil
		case msg := <-c.msgack:
			msg.WaitTimeout(c.republishBackoff, c.republish, c.acktomb.Dying())
		}
	}
}

func (c *msgchan) process(msg *common.Message) {
	if msg.QOS == 0 {
		c.publish(*msg)
		return
	}
	msg.SetCallbackSID(c.persist)
	if !msg.Barrier {
		if msg.TargetQOS == 1 {
			msg.SetAcknowledge()
		}
		c.publish(*msg)
	}
	select {
	case c.msgack <- &common.MsgAck{Message: msg, FST: time.Now()}:
	case <-c.acktomb.Dying():
		return
	}
}

func (c *msgchan) discard(msg *common.Message) {
	c.log.Debugf(errMsgDiscarded.Error())
	if msg.QOS == 0 {
		return
	}
	msg.SetCallbackSID(c.persist)
	select {
	case c.msgack <- &common.MsgAck{Message: msg, FST: time.Now()}:
	case <-c.acktomb.Dying():
		return
	}
}
