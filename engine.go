package baetyl

import (
	"context"
	"errors"
	"io"

	"github.com/baetyl/baetyl/schema/v3"
)

// EngineFactory is the factory method of Engine
type EngineFactory func(Context) (Engine, error)

const defaultEngine = "docker"

var efs map[string]EngineFactory

func init() {
	efs = make(map[string]EngineFactory)
}

// Engine interface
type Engine interface {
	io.Closer
	Name() string
	OSType() string
	Apply(ctx context.Context, appcfg *schema.ComposeAppConfig) error
	UpdateStats()
}

func (rt *runtime) startEngine() error {
	if factory, ok := efs[defaultEngine]; ok {
		e, err := factory(rt)
		if err != nil {
			return err
		}
		rt.e = e
		rt.log.Infoln("engine started")
		return nil
	}
	return errors.New("no such engine")
}

func (rt *runtime) stopEngine() {
	rt.e.Close()
	rt.log.Infoln("engine stopped")
}

// RegisterEngine by name
func RegisterEngine(name string, eh EngineFactory) EngineFactory {
	oeh := efs[name]
	efs[name] = eh
	return oeh
}
