package native

import (
	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
)

// NAME of this engine
const NAME = "native"

func init() {
	engine.Factories()[NAME] = New
}

type nativeEngine struct {
	wdir string
	log  openedge.Logger
}

// Close engine
func (e *nativeEngine) Close() error {
	return nil
}

// Name of engine
func (e *nativeEngine) Name() string {
	return NAME
}

// New native engine
func New(wdir string) (engine.Engine, error) {
	return &nativeEngine{
		wdir: wdir,
		log:  openedge.WithField("engine", NAME),
	}, nil
}
