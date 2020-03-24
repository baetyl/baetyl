package event

import "github.com/baetyl/baetyl-go/faas"

// all event topics
const (
	SyncDesireEvent = "$sync/desire"
	SyncReportEvent = "$sync/report"
	EngineAppEvent  = "$engine/app"
)

type Event struct {
	faas.Message
}

func NewEvent(topic string, payload []byte) *Event {
	return &Event{Message: faas.Message{
		Metadata: map[string]string{"topic": topic},
		Payload:  payload,
	}}
}
