package broker

import (
	"fmt"
	"time"

	"github.com/baidu/openedge/logger"

	"github.com/baidu/openedge/openedge-hub/common"
	"github.com/baidu/openedge/openedge-hub/config"
	"github.com/baidu/openedge/openedge-hub/persist"
	"github.com/baidu/openedge/openedge-hub/utils"
)

var errBrokerClosed = fmt.Errorf("broker alreay closed")

// Offset sink's offset to persist
type Offset struct {
	id    string
	value uint64
}

// Broker a mqtt broker
type Broker struct {
	// message published with qos=0
	msgQ0Chan chan *common.Message
	// message published with qos=1
	msgQ1Chan  chan *common.Message
	msgQ1DB    persist.Database
	offsetDB   persist.Database
	offsetChan chan *Offset
	// others
	config *config.Config
	tomb   utils.Tomb
	log    logger.Logger
}

// NewBroker NewBroker
func NewBroker(c *config.Config, pf *persist.Factory) (b *Broker, err error) {
	msgqos1DB, err := pf.NewDB("msgqos1.db")
	if err != nil {
		return nil, err
	}
	offsetDB, err := pf.NewDB("offset.db")
	if err != nil {
		return nil, err
	}
	b = &Broker{
		config:     c,
		msgQ0Chan:  make(chan *common.Message, c.Message.Ingress.Qos0.Buffer.Size),
		msgQ1Chan:  make(chan *common.Message, c.Message.Ingress.Qos1.Buffer.Size),
		msgQ1DB:    msgqos1DB,
		offsetDB:   offsetDB,
		offsetChan: make(chan *Offset, c.Message.Offset.Buffer.Size),
		log:        logger.WithField("broker", "mqtt"),
	}
	if c.Status.Logging.Enable {
		return b, b.tomb.Gos(b.persistingMsgQos1, b.persistingOffset, b.cleaningMsgQos1, b.logging)
	}
	return b, b.tomb.Gos(b.persistingMsgQos1, b.persistingOffset, b.cleaningMsgQos1)
}

// Config returns config
func (b *Broker) Config() *config.Config {
	return b.config
}

// MsgQ0Chan returns config
func (b *Broker) MsgQ0Chan() <-chan *common.Message {
	return b.msgQ0Chan
}

// Flow flows message to broker
func (b *Broker) Flow(msg *common.Message) {
	b.log.Debugf("flow message: %v", msg)
	if msg.QOS > 0 {
		select {
		case b.msgQ1Chan <- msg:
		case <-b.tomb.Dying():
			b.log.Debugf("flow message (qos=1, pid=%d) failed since broker closed", msg.PacketID)
		}
	} else {
		select {
		case b.msgQ0Chan <- msg:
		case <-b.tomb.Dying():
			b.log.Debugf("flow message (qos=0) failed since broker closed")
		}
	}
}

// FetchQ1 fetches messages with qos=1
func (b *Broker) FetchQ1(offset uint64, batchSize int) ([]*common.Message, error) {
	if !b.tomb.Alive() {
		return nil, errBrokerClosed
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
func (b *Broker) OffsetChanLen() int {
	return len(b.offsetChan)
}

// OffsetPersisted gets sink's offset from database
func (b *Broker) OffsetPersisted(id string) (*uint64, error) {
	if !b.tomb.Alive() {
		return nil, errBrokerClosed
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
func (b *Broker) WaitOffsetPersisted() {
	for {
		if len(b.offsetChan) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// PersistOffset puts sink's offset to offset channal
func (b *Broker) PersistOffset(id string, offset uint64) error {
	if !b.tomb.Alive() {
		return errBrokerClosed
	}
	b.offsetChan <- &Offset{id, offset}
	return nil
}

// DeleteOffset  delete sink's offset from database
func (b *Broker) deleteOffset(id string) error {
	if !b.tomb.Alive() {
		return errBrokerClosed
	}
	err := b.offsetDB.Delete([]byte(id))
	if err != nil {
		return err
	}
	return nil
}

func (b *Broker) sequence() (uint64, error) {
	if !b.tomb.Alive() {
		return 0, errBrokerClosed
	}
	return b.msgQ1DB.Sequence()
}

// InitOffset init offset
func (b *Broker) InitOffset(id string, persistent bool) (uint64, error) {
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

// Close closes broker
func (b *Broker) Close() {
	b.log.Infof("broker closing")
	b.tomb.Kill()
	err := b.tomb.Wait()
	b.log.WithError(err).Infof("broker closed")
}
