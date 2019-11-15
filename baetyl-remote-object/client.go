package main

import (
	"io"
)

// CallAsync interface
type CallAsync func(msg *EventMessage, cb func(msg *EventMessage, err error)) error

// Start interface
type Start func() error

// Report reports stats
type report func(map[string]interface{}) error

// Client interface
type Client interface {
	CallAsync(msg *EventMessage, cb func(msg *EventMessage, err error)) error
	Start() error
	io.Closer
}

// NewClient can create a ruler
func NewClient(cfg ClientInfo, r report) (Client, error) {
	return NewStorageClient(cfg, r)
}
