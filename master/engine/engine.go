package engine

import (
	"errors"
	"io"
	"time"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// Factory create engine by given config
type Factory func(is InfoStats, opts Options) (Engine, error)

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
	Prepare(baetyl.ComposeAppConfig)
	SetInstanceStats(serviceName, instanceName string, partialStats PartialStats, persist bool)
	DelInstanceStats(serviceName, instanceName string, persist bool)
	DelServiceStats(serviceName string, persist bool)
	Run(string, baetyl.ComposeService, map[string]baetyl.ComposeVolume) (Service, error)
}

// Options engine options
type Options struct {
	Name string
	Grace time.Duration
	Pwd string
	APIVersion string
}

// New engine by given name
func New(name string, is InfoStats, opts Options) (Engine, error) {
	if f, ok := factories[name]; ok {
		return f(is, opts)
	}
	return nil, errors.New("no such engine")
}
