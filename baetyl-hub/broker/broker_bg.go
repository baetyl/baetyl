package broker

import (
	"time"

	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/baetyl-hub/persist"
	"github.com/baetyl/baetyl/baetyl-hub/utils"
	"github.com/golang/protobuf/proto"
)

func (b *Broker) reporting() error {
	defer b.log.Debugf("metrics reporting task stopped")

	var err error
	t := time.NewTicker(b.config.Metrics.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-b.tomb.Dying():
			return nil
		case <-t.C:
			stats := map[string]interface{}{
				"broker_stats": map[string]interface{}{
					"ingress_buffered_message_qos0":  len(b.msgQ0Chan),
					"ingress_buffered_message_qos1":  len(b.msgQ1Chan),
					"egress_buffered_offset": len(b.offsetChan),
				},
			}
			err = b.report(stats)
			if err != nil {
				b.log.Warnf("failed to report broker stats")
			}
			b.log.Debugln(stats)
		}
	}
}

func (b *Broker) cleaningMsgQos1() error {
	defer b.log.Debugf("cleaning message (qos=1) task stopped")

	retention := b.config.Message.Ingress.Qos1.Cleanup.Retention
	t := time.NewTicker(b.config.Message.Ingress.Qos1.Cleanup.Interval)
	defer t.Stop()
	for {
		select {
		case <-b.tomb.Dying():
			return nil
		case <-t.C:
			start := time.Now()
			timestamp := uint64(start.Unix()) - uint64(retention.Seconds())
			c, err := b.msgQ1DB.Clean(timestamp)
			elapsed := time.Since(start)
			if c > 0 {
				b.log.Infof("cleanup %d message(s): elapsed=%v, timestamp=%d, error=%v", c, elapsed, timestamp, err)
			}
		}
	}
}

func (b *Broker) persistingMsgQos1() error {
	defer b.log.Debugf("persisting message (qos=1) task stopped")
	msgs := make([]*common.Message, 0)
	batchSize := b.config.Message.Ingress.Qos1.Buffer.Size
	ticker := time.NewTicker(time.Millisecond * 10)
loop:
	for {
		select {
		case <-b.tomb.Dying():
			break loop
		case m := <-b.msgQ1Chan:
			msgs = append(msgs, m)
			if len(msgs) >= batchSize {
				msgs = b.persistMessages(msgs)
			}
		case <-ticker.C:
			if len(msgs) > 0 {
				msgs = b.persistMessages(msgs)
			}
		}
	}
	// Try to persist all messages
last_loop:
	for {
		select {
		case m := <-b.msgQ1Chan:
			msgs = append(msgs, m)
		default:
			break last_loop
		}
	}
	if len(msgs) > 0 {
		b.persistMessages(msgs)
	}
	return nil
}

func (b *Broker) persistingOffset() error {
	defer b.log.Debugf("persisting offset task stopped")

	count := 0
	offsets := make(map[string]uint64)
	batchSize := b.config.Message.Offset.Batch.Max
	ticker := time.NewTicker(time.Millisecond * 100)
loop:
	for {
		select {
		case <-b.tomb.Dying():
			break loop
		case o := <-b.offsetChan:
			count++
			offsets[o.id] = o.value
			if count >= batchSize {
				offsets = b.persistOffsets(offsets)
				count = 0
			}
		case <-ticker.C:
			if len(offsets) > 0 {
				offsets = b.persistOffsets(offsets)
				count = 0
			}
		}
	}
	// Try to persist all offsets
last_loop:
	for {
		select {
		case o := <-b.offsetChan:
			offsets[o.id] = o.value
		default:
			break last_loop
		}
	}
	if len(offsets) > 0 {
		b.persistOffsets(offsets)
	}
	return nil
}

func (b *Broker) persistMessages(msgs []*common.Message) (empty []*common.Message) {
	l := len(msgs)
	empty = make([]*common.Message, 0)
	if l == 0 {
		return
	}
	var err error
	vs := make([][]byte, l)
	for i, m := range msgs {
		vs[i], err = proto.Marshal(&m.Persisted)
		if err != nil {
			b.log.WithError(err).Errorf("failed to marshal persisted message")
			return
		}
	}
	err = b.msgQ1DB.BatchPutV(vs)
	if err != nil {
		b.log.WithError(err).Errorf("failed to persist %d message(s)", l)
		return
	}
	b.log.Debugf("%d message(s) persisted", l)
	for _, msg := range msgs {
		msg.CallbackPID()
	}
	return
}

func (b *Broker) persistOffsets(offsets map[string]uint64) (empty map[string]uint64) {
	empty = make(map[string]uint64)
	kvs := make([]*persist.KV, 0)
	for k, v := range offsets {
		kvs = append(kvs, &persist.KV{Key: []byte(k), Value: utils.U64ToB(v)})
	}
	err := b.offsetDB.BatchPut(kvs)
	if err != nil {
		b.log.WithError(err).Errorf("failed to persist %d offset(s)", len(offsets))
	} else {
		b.log.Debugf("%d offset(s) persisted: %v", len(offsets), offsets)
	}
	return
}
