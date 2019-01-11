package main

import (
	"encoding/json"
	"time"
)

// EventType the type of event from cloud
type EventType string

// EventHandle the handler of event from cloud
type EventHandle func(e *Event)

// The type of event from cloud
const (
	Start      EventType = "START"
	Stop       EventType = "STOP"
	SyncConfig EventType = "SYNC_CONFIG"
	Unkown     EventType = "UNKNOWN"
)

// Detail event details
type Detail struct {
	DownloadURL string `json:"downloadUri,omitempty"`
	Version     string `json:"version,omitempty"`
}

// Event event message
type Event struct {
	Time   time.Time `json:"time"`
	Type   EventType `json:"event"`
	Detail Detail    `json:"detail"`
}

// NewEvent ceates a new event
func NewEvent(v []byte) *Event {
	var e Event
	json.Unmarshal(v, &e)
	return &e
}

// Bytes masrshal event to json string
func (e *Event) Bytes() []byte {
	v, _ := json.Marshal(e)
	return v
}

func newStartEvent(v string) *Event {
	return &Event{
		Time:   time.Now(),
		Type:   Start,
		Detail: Detail{Version: v},
	}
}

func newStopEvent() *Event {
	return &Event{
		Time:   time.Now(),
		Type:   Stop,
		Detail: Detail{},
	}
}
