package engine

import (
	"errors"
	"io"
	"time"

	"github.com/baidu/openedge/sdk-go/openedge"
)

// Factory create engine by given config
type Factory func(grace time.Duration, wdir string) (Engine, error)

var factories map[string]Factory

func init() {
	factories = make(map[string]Factory)
}

// Factories of engines
func Factories() map[string]Factory {
	return factories
}

// Engine interface
type Engine interface {
	io.Closer
	Name() string
	Run(openedge.ServiceInfo, map[string]openedge.VolumeInfo) (Service, error)
}

// New engine by given name
func New(name string, grace time.Duration, wdir string) (Engine, error) {
	if f, ok := factories[name]; ok {
		return f(grace, wdir)
	}
	return nil, errors.New("no such engine")
}
