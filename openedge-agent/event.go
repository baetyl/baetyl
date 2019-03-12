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
