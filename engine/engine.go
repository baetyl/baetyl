package engine

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/event"
	"os"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

type Engine struct {
	sha   *shadow.Shadow
	cent  *event.Center
	cfg   config.EngineConfig
	model ami.Model
	tomb  utils.Tomb
	log   *log.Logger
}

func NewEngine(cfg config.EngineConfig, model ami.Model, sha *shadow.Shadow, cent *event.Center) (*Engine, error) {
	if cfg.Kind != "kubernetes" {
		return nil, os.ErrInvalid
	}
	e := &Engine{
		sha:   sha,
		model: model,
		cent:  cent,
		cfg:   cfg,
		log:   log.With(log.Any("engine", cfg.Kind)),
	}
	e.tomb.Go(e.collecting)
	return e, nil
}

func (e *Engine) collecting() error {
	t := time.NewTicker(e.cfg.Collector.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			info, err := e.model.CollectInfo()
			if err != nil {
				e.log.Error("failed to collect info", log.Error(err))
				continue
			}
			delta, err := e.sha.Report(info)
			if err != nil {
				e.log.Error("failed to update shadow report", log.Error(err))
				continue
			}
			var msg link.Message
			if len(delta) > 0 {
				msg.Content, err = json.Marshal(delta)
				if err != nil {
					e.log.Error("failed to marshal delta", log.Error(err))
					continue
				}
				msg.Context.Topic = common.SyncDesireEvent
			} else {
				msg.Content, err = json.Marshal(info)
				if err != nil {
					e.log.Error("failed to marshal delta", log.Error(err))
					continue
				}
				msg.Context.Topic = common.SyncReportEvent
			}
			err = e.cent.Trigger(&msg)
			if err != nil {
				e.log.Error("failed to trigger event", log.Error(err))
				continue
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

func (e *Engine) Apply(msg link.Message) error {
	var info map[string]interface{}
	err := json.Unmarshal(msg.Content, &info)
	if err != nil {
		return err
	}
	apps, ok := info["apps"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("apps does not exist")
	}
	err = e.model.ApplyApplications(apps)
	if err != nil {
		return err
	}
	return nil
}
