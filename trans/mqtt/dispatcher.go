package mqtt

import (
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/utils"
	"github.com/jpillora/backoff"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
)

// ErrDispatcherClosed is returned if the dispatcher is closed
var ErrDispatcherClosed = errors.New("dispatcher closed")

// Dispatcher dispatcher of mqtt client
type Dispatcher struct {
	config  ClientConfig
	channel chan packet.Generic
	backoff *backoff.Backoff
	tomb    utils.Tomb
	log     *logrus.Entry
}

// NewDispatcher creata a new dispatcher
func NewDispatcher(cc ClientConfig) *Dispatcher {
	if cc.Address == "" {
		return nil
	}
	return &Dispatcher{
		config:  cc,
		channel: make(chan packet.Generic, cc.BufferSize),
		backoff: &backoff.Backoff{
			Min:    time.Millisecond * 500,
			Max:    cc.Interval,
			Factor: 2,
		},
		log: logger.WithFields("dispatcher", "mqtt"),
	}
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
func (d *Dispatcher) Start(cb func(packet.Generic)) error {
	return d.tomb.Go(func() error {
		return d.supervisor(cb)
	})
}

// Close closes dispatcher
func (d *Dispatcher) Close() error {
	d.tomb.Kill(nil)
	return d.tomb.Wait()
}

// Supervisor the supervised reconnect loop
func (d *Dispatcher) supervisor(cb func(packet.Generic)) error {
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

			d.log.Debugln("Delay reconnect:", next)

			// sleep but return on Stop
			select {
			case <-time.After(next):
			case <-d.tomb.Dying():
				return nil
			}
		}

		d.log.Debugln("Next reconnect")

		// prepare the stop channel
		fail := make(chan struct{})

		// try once to get a client
		callback := func(pkt packet.Generic, err error) {
			if err != nil {
				close(fail)
				return
			}
			if cb != nil {
				cb(pkt)
			}
		}
		client, err := NewClient(d.config, callback)
		if err != nil {
			d.log.WithError(err).Errorln("Failed to create new client")
			continue
		}

		// run callback
		d.log.Debugln("Client online")

		// run dispatcher on client
		current, dying = d.dispatcher(client, current, fail)

		// run callback
		d.log.Debugln("Client offline")

		// return goroutine if dying
		if dying {
			return nil
		}
	}
}

// reads from the queues and calls the current client
func (d *Dispatcher) dispatcher(client *Client, current packet.Generic, fail chan struct{}) (packet.Generic, bool) {
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
		case <-d.tomb.Dying():
			return nil, true
		case <-fail:
			return nil, false
		}
	}
}
