package engine

import (
	"errors"

	openedge "github.com/baidu/openedge/api/go"
)

// Factory create engine by given config
type Factory func(wdir string) (Engine, error)

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
	Close() error
	Name() string
	Run(name string, service *openedge.ServiceInfo) (Service, error)
	RunWithConfig(name string, service *openedge.ServiceInfo, config []byte) (Service, error)
}

// New engine by given name
func New(name string, wdir string) (Engine, error) {
	if f, ok := factories[name]; ok {
		return f(wdir)
	}
	return nil, errors.New("no such engine")
}
