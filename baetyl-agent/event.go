package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/256dpi/gomqtt/packet"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// EventType the type of event from cloud
type EventType string

// EventHandle the handler of event from cloud
type EventHandle func(e *Event)

// The type of event from cloud
const (
	OTA     EventType = "OTA"
	Update  EventType = "UPDATE" // deprecated
	Unknown EventType = "UNKNOWN"
)

// Event event message
type Event struct {
	Time    time.Time   `json:"time"`
	Type    EventType   `json:"event"`
	Content interface{} `json:"content"`
}

// NewEvent creates a new event
func NewEvent(v []byte) (*Event, error) {
	e := &Event{}
	err := json.Unmarshal(v, e)
	if err != nil {
		return nil, fmt.Errorf("event invalid: %s", err.Error())
	}
	switch e.Type {
	case OTA:
		e.Content = &EventOTA{}
	case Update:
		e.Content = &UpdateEvent{}
	default:
		return nil, fmt.Errorf("event type (%s) unexpected", e.Type)
	}
	err = json.Unmarshal(v, e)
	if err != nil {
		return nil, fmt.Errorf("event content invalid: %s", err.Error())
	}
	return e, nil
}

// EventOTA OTA event
type EventOTA struct {
	Type    string            `json:"type,omitempty"`
	Trace   string            `json:"trace,omitempty"`
	Version string            `json:"version,omitempty"`
	Volume  baetyl.VolumeInfo `json:"volume,omitempty"`
}

// UpdateEvent update event
// TODO: deprecate
type UpdateEvent struct {
	Trace   string            `json:"trace,omitempty"`
	Version string            `json:"version,omitempty"`
	Config  baetyl.VolumeInfo `json:"config,omitempty"`
}

func (a *agent) ProcessPublish(p *packet.Publish) error {
	if p.Message.QOS == 1 {
		puback := packet.NewPuback()
		puback.ID = p.ID
		a.mqtt.Send(puback)
	}
	e, err := NewEvent(p.Message.Payload)
	if err != nil {
		return err
	}
	// convert update event to ota event
	ue, ok := e.Content.(*UpdateEvent)
	if ok {
		e.Type = OTA
		e.Content = &EventOTA{
			Type:    baetyl.OTAAPP,
			Trace:   ue.Trace,
			Version: ue.Version,
			Volume:  ue.Config,
		}
	}
	a.ctx.Log().Debugln("event:", string(p.Message.Payload))
	select {
	case oe := <-a.events:
		a.ctx.Log().Warnf("discard old event: %+v", *oe)
		a.events <- e
	case a.events <- e:
	case <-a.tomb.Dying():
	}
	return nil
}

func (a *agent) ProcessPuback(p *packet.Puback) error {
	return nil
}

func (a *agent) ProcessError(err error) {
	a.ctx.Log().Errorf(err.Error())
}
