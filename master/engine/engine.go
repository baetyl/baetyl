package engine

import (
	"errors"
	"io"
	"time"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// Factory create engine by given config
type Factory func(grace time.Duration, pwd string, is InfoStats) (Engine, error)

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
	Recover()
	Prepare([]openedge.ServiceInfo)
	SetInstanceStats(serviceName, instanceName string, partialStats PartialStats, persist bool)
	DelInstanceStats(serviceName, instanceName string, persist bool)
	DelServiceStats(serviceName string, persist bool)
	Run(openedge.ServiceInfo, map[string]openedge.VolumeInfo) (Service, error)
}

// New engine by given name
func New(name string, grace time.Duration, pwd string, is InfoStats) (Engine, error) {
	if f, ok := factories[name]; ok {
		return f(grace, pwd, is)
	}
	return nil, errors.New("no such engine")
}
