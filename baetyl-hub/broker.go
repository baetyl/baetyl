package hub

import (
	"fmt"
	"time"

	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/baetyl-hub/config"
	"github.com/baetyl/baetyl/baetyl-hub/utils"
	"github.com/baetyl/baetyl/logger"
)

var errbrokerClosed = fmt.Errorf("broker already closed")

type offset struct {
	id    string
	value uint64
}

// broker a mqtt broker
type broker struct {
	// message published with qos=0
	msgQ0Chan chan *common.Message
	// message published with qos=1
	msgQ1Chan  chan *common.Message
	msgQ1DB    Database
	offsetDB   Database
	offsetChan chan *offset
	// others
	config *config.Config
	tomb   utils.Tomb
	log    logger.Logger
}

func (h *hub) startBroker() error {
	c := &h.cfg
	pf := h.storage
	msgqos1DB, err := pf.newDB("msgqos1.db")
	if err != nil {
		return err
	}
	offsetDB, err := pf.newDB("offset.db")
	if err != nil {
		return err
	}
	b := &broker{
		config:     c,
		msgQ0Chan:  make(chan *common.Message, c.Message.Ingress.Qos0.Buffer.Size),
		msgQ1Chan:  make(chan *common.Message, c.Message.Ingress.Qos1.Buffer.Size),
		msgQ1DB:    msgqos1DB,
		offsetDB:   offsetDB,
		offsetChan: make(chan *offset, c.Message.Offset.Buffer.Size),
		log:        logger.WithField("broker", "mqtt"),
	}
	h.broker = b
	return b.tomb.Gos(b.persistingMsgQos1, b.persistingOffset, b.cleaningMsgQos1)
}

// Config returns config
func (b *broker) Config() *config.Config {
	return b.config
}

// MsgQ0Chan returns config
func (b *broker) MsgQ0Chan() <-chan *common.Message {
	return b.msgQ0Chan
}

// Flow flows message to broker
func (b *broker) Flow(msg *common.Message) {
	b.log.Debugf("flow message: %v", msg)
	if msg.QOS > 0 {
		select {
		case b.msgQ1Chan <- msg:
		case <-b.tomb.Dying():
			b.log.Debugf("failed to flow message (qos=1, pid=%d) since broker closed", msg.PacketID)
		}
	} else {
		select {
		case b.msgQ0Chan <- msg:
		case <-b.tomb.Dying():
			b.log.Debugf("failed to flow message (qos=0) since broker closed")
		}
	}
}

// FetchQ1 fetches messages with qos=1
func (b *broker) FetchQ1(offset uint64, batchSize int) ([]*common.Message, error) {
	if !b.tomb.Alive() {
		return nil, errbrokerClosed
	}
	kvs, err := b.msgQ1DB.BatchFetch(utils.U64ToB(offset), batchSize)
	if err != nil {
		return nil, err
	}
	msgs := make([]*common.Message, len(kvs))
	for i, kv := range kvs {
		msg, err := common.UnmarshalMessage(kv.Key, kv.Value)
		if err != nil {
			return nil, err
		}
		msgs[i] = msg
	}
	return msgs, nil
}

// OffsetChanLen returns the length of offset channel
func (b *broker) OffsetChanLen() int {
	return len(b.offsetChan)
}

// OffsetPersisted gets sink's offset from database
func (b *broker) OffsetPersisted(id string) (*uint64, error) {
	if !b.tomb.Alive() {
		return nil, errbrokerClosed
	}
	v, err := b.offsetDB.Get([]byte(id))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	res := utils.U64(v)
	return &res, nil
}

// WaitOffsetPersisted waits all offsets in channel to be persisted
func (b *broker) WaitOffsetPersisted() {
	for {
		if len(b.offsetChan) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// PersistOffset puts sink's offset to offset channal
func (b *broker) PersistOffset(id string, off uint64) error {
	if !b.tomb.Alive() {
		return errbrokerClosed
	}
	b.offsetChan <- &offset{id, off}
	return nil
}

// DeleteOffset  delete sink's offset from database
func (b *broker) deleteOffset(id string) error {
	if !b.tomb.Alive() {
		return errbrokerClosed
	}
	err := b.offsetDB.Delete([]byte(id))
	if err != nil {
		return err
	}
	return nil
}

func (b *broker) sequence() (uint64, error) {
	if !b.tomb.Alive() {
		return 0, errbrokerClosed
	}
	return b.msgQ1DB.Sequence()
}

// InitOffset init offset
func (b *broker) InitOffset(id string, persistent bool) (uint64, error) {
	offset, err := b.sequence()
	if err != nil {
		return 0, err
	}
	if persistent {
		// get old offset if exists
		v, err := b.OffsetPersisted(id)
		if err != nil {
			return 0, err
		}
		if v == nil {
			// persist offset at first time
			err = b.PersistOffset(id, offset)
			if err != nil {
				return 0, err
			}
		} else {
			offset = *v
		}
	} else {
		// delete old offset if exists
		err := b.deleteOffset(id)
		if err != nil {
			return 0, err
		}
	}
	return offset + 1, nil
}

func (h *hub) stopBroker() {
	h.broker.log.Infof("broker closing")
	h.broker.tomb.Kill()
	err := h.broker.tomb.Wait()
	h.broker.log.WithError(err).Infof("broker closed")
}
