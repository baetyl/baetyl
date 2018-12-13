package common

import (
	"time"
)

// Subscribe MQTT subscribe
type Subscribe struct {
	sequenceID  uint64
	acknowledge *Acknowledge
}

// NewSubscribe creates a subscribe
func NewSubscribe() *Subscribe {
	return &Subscribe{
		sequenceID:  uint64(time.Now().UnixNano()),
		acknowledge: NewAcknowledge(),
	}
}

// Ack acknowledge
func (s *Subscribe) Ack() {
	s.acknowledge.Ack()
}

// SID sequence id
func (s *Subscribe) SID() uint64 {
	return s.sequenceID
}

// WaitTimeout waits until acknowledged, cancelled or timeout
func (s *Subscribe) WaitTimeout(timeout time.Duration, cancel <-chan struct{}) bool {
	select {
	case <-cancel:
		return false
	case <-time.After(timeout):
		return false
	case <-s.acknowledge.done:
		return true
	}
}
