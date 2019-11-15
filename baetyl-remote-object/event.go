package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventType the type of event from cloud
type EventType string

// The type of event from cloud
const (
	Upload  EventType = "UPLOAD"
	Package EventType = "PACKAGE"
)

// Event event message
type Event struct {
	Time    time.Time   `json:"time"`
	Type    EventType   `json:"type"`
	Content interface{} `json:"content"`
}

// EventMessage config
type EventMessage struct {
	ID    uint64
	QOS   uint32
	Topic string
	Event *Event
}

// NewEvent creates a new event
func NewEvent(v []byte) (*Event, error) {
	var e Event
	json.Unmarshal(v, &e)
	switch e.Type {
	case Upload:
		e.Content = &UploadEvent{}
		json.Unmarshal(v, &e)
		return &e, nil
	default:
		return nil, fmt.Errorf("event type unexpected")
	}
}

// UploadEvent update event
type UploadEvent struct {
	RemotePath string            `yaml:"remotePath" json:"remotePath" validate:"nonzero"`
	LocalPath  string            `yaml:"localPath" json:"localPath" validate:"nonzero"`
	Zip        bool              `yaml:"zip" json:"zip"`
	Meta       map[string]string `yaml:"meta" json:"meta" default:"{}"`
}
