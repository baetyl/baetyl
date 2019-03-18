package main

import (
	"encoding/json"
	"time"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// EventType the type of event from cloud
type EventType string

// EventHandle the handler of event from cloud
type EventHandle func(e *Event)

// The type of event from cloud
const (
	Update  EventType = "UPDATE"
	Unknown EventType = "UNKNOWN"
)

// Event event message
type Event struct {
	Time    time.Time `json:"time"`
	Type    EventType `json:"event"`
	Content []byte    `json:"content"`
}

// NewEvent creates a new event
func NewEvent(v []byte) *Event {
	var e Event
	json.Unmarshal(v, &e)
	return &e
}

// UpdateEvent update event
type UpdateEvent struct {
	Version string              `yaml:"version" json:"version"`
	Clean   bool                `yaml:"clean" json:"clean"`
	Config  openedge.VolumeInfo `yaml:"config" json:"config"`
}

func newUpdateEvent(d []byte) (*UpdateEvent, error) {
	data := new(UpdateEvent)
	err := utils.UnmarshalJSON(d, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
