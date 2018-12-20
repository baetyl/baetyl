package main

import (
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/function/runtime"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
)

// ErrDispatcherClosed is returned if the dispatcher is closed
var ErrDispatcherClosed = fmt.Errorf("dispatcher closed")

// Dispatcher dispatcher of mqtt client
type Dispatcher struct {
	function *Function
	callback func(*packet.Publish)
	buffer   chan struct{}
	tomb     utils.Tomb
	log      *logger.Entry
}

// NewDispatcher creata a new dispatcher
func NewDispatcher(f *Function) (*Dispatcher, error) {
	return &Dispatcher{
		function: f,
		buffer:   make(chan struct{}, f.cfg.Instance.Max),
		log:      logger.WithFields("dispatcher", "func"),
	}, nil
}

// SetCallback sets callback
func (d *Dispatcher) SetCallback(c func(p *packet.Publish)) {
	d.callback = c
}

// Invoke invokes a function
func (d *Dispatcher) Invoke(pkt *packet.Publish) error {
	select {
	case d.buffer <- struct{}{}:
	case <-d.tomb.Dying():
		return ErrDispatcherClosed
	}
	go func(pub *packet.Publish) {
		msg := &runtime.Message{
			QOS:     uint32(pub.Message.QOS),
			Topic:   pub.Message.Topic,
			Payload: pub.Message.Payload,
		}
		msg, err := d.function.Invoke(msg)
		if err != nil {
			pub.Message.Payload = utils.MakeErrorPayload(pub, err)
		} else {
			pub.Message.Payload = msg.Payload
		}
		if d.callback != nil {
			d.callback(pub)
		}
		<-d.buffer
	}(pkt)
	return nil
}

// Close closes dispatcher
func (d *Dispatcher) Close() error {
	defer d.log.Debugf("function dispatcher closed")
	
	d.tomb.Kill(nil)
	return d.tomb.Wait()
}
