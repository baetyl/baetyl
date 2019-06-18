package main

import (
	"encoding/json"
	"fmt"
	"time"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
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
	Time    time.Time   `json:"time"`
	Type    EventType   `json:"event"`
	Content interface{} `json:"content"`
}

// NewEvent creates a new event
func NewEvent(v []byte) (*Event, error) {
	var e Event
	json.Unmarshal(v, &e)
	switch e.Type {
	case Update:
		e.Content = &UpdateEvent{}
		json.Unmarshal(v, &e)
		return &e, nil
	default:
		return nil, fmt.Errorf("event type unexpected")
	}
}

// UpdateEvent update event
type UpdateEvent struct {
	Version string              `yaml:"version" json:"version"`
	Config  openedge.VolumeInfo `yaml:"config" json:"config"`
}
