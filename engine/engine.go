package engine

import (
	"encoding/json"
	"github.com/baetyl/baetyl-core/omi"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"time"
)

type Engine struct {
	t     *time.Ticker
	tomb  utils.Tomb
	model omi.Model
	log   *log.Logger
}

func NewEngine(interval time.Duration, model omi.Model, log *log.Logger) *Engine {
	e := &Engine{
		log:   log,
		model: model,
		t:     time.NewTicker(interval),
	}
	e.tomb.Go(e.start)
	return e
}

func (e *Engine) start() error {
	for {
		select {
		case <-e.t.C:
			res, err := e.model.CollectInfo(map[string]string{})
			if err != nil {
				e.log.Error("failed to collect info", log.Error(err))
			}
			err = e.updateShadowReport(res)
			if err != nil {
				e.log.Error("failed to update shadow report", log.Error(err))
			}
		case <-e.tomb.Dying():
			return nil
		}
	}
}

func (e *Engine) Close() {
	e.tomb.Kill(nil)
	e.tomb.Wait()
}

func (e *Engine) updateShadowReport(res interface{}) error {
	// TODO update shadow report
	return nil
}

func (e *Engine) Handler(msg link.Message) error {
	var apps map[string]string
	err := json.Unmarshal(msg.Content, &apps)
	if err != nil {
		return err
	}
	err = e.model.ApplyApplications(apps)
	if err != nil {
		return err
	}
	return nil
}
