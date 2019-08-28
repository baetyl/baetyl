package mqtt

import (
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
	"github.com/jpillora/backoff"
)

// ErrDispatcherClosed is returned if the dispatcher is closed
var ErrDispatcherClosed = fmt.Errorf("dispatcher already closed")

// Dispatcher dispatcher of mqtt client
type Dispatcher struct {
	config  ClientInfo
	channel chan packet.Generic
	backoff *backoff.Backoff
	tomb    utils.Tomb
	log     logger.Logger
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(cc ClientInfo, log logger.Logger) *Dispatcher {
	if log == nil {
		log = logger.Global
	}
	return &Dispatcher{
		config:  cc,
		channel: make(chan packet.Generic, cc.BufferSize),
		backoff: &backoff.Backoff{
			Min:    time.Millisecond * 500,
			Max:    cc.Interval,
			Factor: 2,
		},
		log: log.WithField("mqtt", "dispatcher").WithField("cid", cc.ClientID),
	}
}

// Publish sends a publish packet
func (d *Dispatcher) Publish(pid uint16, qos uint32, topic string, payload []byte, retain bool, duplicate bool) error {
	pkt := packet.NewPublish()
	pkt.ID = packet.ID(pid)
	pkt.Dup = duplicate
	pkt.Message.QOS = packet.QOS(qos)
	pkt.Message.Topic = topic
	pkt.Message.Payload = payload
	pkt.Message.Retain = retain
	return d.Send(pkt)
}

// Send sends a generic packet
func (d *Dispatcher) Send(pkt packet.Generic) error {
	select {
	case d.channel <- pkt:
	case <-d.tomb.Dying():
		return ErrDispatcherClosed
	}
	return nil
}

// Start starts dispatcher
func (d *Dispatcher) Start(h Handler) error {
	return d.tomb.Go(func() error {
		return d.supervisor(h)
	})
}

// Close closes dispatcher
func (d *Dispatcher) Close() error {
	d.tomb.Kill(nil)
	return d.tomb.Wait()
}

// Supervisor the supervised reconnect loop
func (d *Dispatcher) supervisor(handler Handler) error {
	first := true
	var dying bool
	var current packet.Generic

	for {
		if first {
			// no delay on first attempt
			first = false
		} else {
			// get backoff duration
			next := d.backoff.Duration()

			d.log.Debugln("delay reconnect:", next)

			// sleep but return on Stop
			select {
			case <-time.After(next):
			case <-d.tomb.Dying():
				return nil
			}
		}

		d.log.Debugln("next reconnect")

		client, err := NewClient(d.config, handler, d.log)
		if err != nil {
			d.log.WithError(err).Errorln("failed to create new client")
			continue
		}

		// run callback
		d.log.Debugln("client online")

		// run dispatcher on client
		current, dying = d.dispatcher(client, current)

		// run callback
		d.log.Debugln("client offline")

		// return goroutine if dying
		if dying {
			return nil
		}
	}
}

// reads from the queues and calls the current client
func (d *Dispatcher) dispatcher(client *Client, current packet.Generic) (packet.Generic, bool) {
	defer client.Close()

	if current != nil {
		err := client.Send(current)
		if err != nil {
			return current, false
		}
	}

	for {
		select {
		case pkt := <-d.channel:
			err := client.Send(pkt)
			if err != nil {
				return pkt, false
			}
		case <-client.Dying():
			return nil, false
		case <-d.tomb.Dying():
			return nil, true
		}
	}
}
