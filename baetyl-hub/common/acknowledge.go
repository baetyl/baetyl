package common

import (
	"sync/atomic"
)

// Acknowledge acknowledgement
type Acknowledge struct {
	count int32
	done  chan struct{}
}

// NewAcknowledge creates a new acknowledgement
func NewAcknowledge() *Acknowledge {
	return &Acknowledge{
		count: 1,
		done:  make(chan struct{}),
	}
}

// Ack acknowledges after message handled
func (ack *Acknowledge) Ack() {
	if atomic.AddInt32(&ack.count, -1) == 0 {
		close(ack.done)
	}
}

// Count returns the ack count
func (ack *Acknowledge) Count() int32 {
	return atomic.LoadInt32(&ack.count)
}

// Wait waits until acknowledged or cancelled
func (ack *Acknowledge) Wait(cancel <-chan struct{}) bool {
	select {
	case <-cancel:
		return false
	case <-ack.done:
		return true
	}
}
