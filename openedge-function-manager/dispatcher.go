package main

import (
	"encoding/json"
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	"github.com/docker/distribution/uuid"
)

// ErrDispatcherClosed is returned if the dispatcher is closed
var ErrDispatcherClosed = fmt.Errorf("dispatcher closed")

// Dispatcher dispatcher of function client
type Dispatcher struct {
	function *Function
	callback func(*packet.Publish)
	buffer   chan struct{}
	tomb     utils.Tomb
	log      logger.Logger
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(ctx openedge.Context, cfg FunctionInfo) *Dispatcher {
	f := NewFunction(ctx, cfg)
	return &Dispatcher{
		function: f,
		buffer:   make(chan struct{}, f.pool.Config.MaxTotal),
		log:      f.log.WithField("dispatcher", "function"),
	}
}

// SetCallback sets callback
func (d *Dispatcher) SetCallback(c func(p *packet.Publish)) {
	d.callback = c
}

// Call calls a function
func (d *Dispatcher) Call(pkt *packet.Publish) error {
	select {
	case d.buffer <- struct{}{}:
	case <-d.tomb.Dying():
		return ErrDispatcherClosed
	}
	go func(p *packet.Publish) {
		msg := &openedge.FunctionMessage{
			QOS:              uint32(p.Message.QOS),
			Topic:            p.Message.Topic,
			Payload:          p.Message.Payload,
			FunctionName:     d.function.cfg.Name,
			FunctionInvokeID: uuid.Generate().String(),
		}
		out, err := d.function.Call(msg)
		// TODO: add retry logic
		if err != nil {
			p.Message.Payload = MakeErrorPayload(p, err)
		} else {
			p.Message.Payload = out.Payload
		}
		if d.callback != nil {
			d.callback(p)
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

// MakeErrorPayload makes error payload
func MakeErrorPayload(p *packet.Publish, err error) []byte {
	s := make(map[string]interface{})
	s["packet"] = p
	s["errorMessage"] = err.Error()
	s["errorType"] = fmt.Sprintf("%T", err)
	o, _ := json.Marshal(s)
	return o
}
