package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/event"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-go/log"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
)

type Engine struct {
	sha  *shadow.Shadow
	cent *event.Center
	cfg  config.EngineConfig
	ami  ami.AMI
	tomb utils.Tomb
	log  *log.Logger
}

func NewEngine(cfg config.EngineConfig, ami ami.AMI, sha *shadow.Shadow, cent *event.Center) (*Engine, error) {
	if cfg.Kind != "kubernetes" {
		return nil, os.ErrInvalid
	}
	e := &Engine{
		sha:  sha,
		ami:  ami,
		cent: cent,
		cfg:  cfg,
		log:  log.With(log.Any("engine", cfg.Kind)),
	}
	return e, nil
}

func (e *Engine) Start() {
	e.tomb.Go(e.collecting)
}

func (e *Engine) collecting() error {
	t := time.NewTicker(e.cfg.Collector.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			e.collect()
		case <-e.tomb.Dying():
			return nil
		}
	}
}

func (e *Engine) collect() {
	info, err := e.ami.Collect()
	if err != nil {
		e.log.Error("failed to collect info", log.Error(err))
		return
	}
	delta, err := e.sha.Report(info)
	if err != nil {
		e.log.Error("failed to update shadow report", log.Error(err))
		return
	}
	var evt *event.Event
	if len(delta) > 0 {
		pld, err := json.Marshal(delta)
		if err != nil {
			e.log.Error("failed to marshal delta", log.Error(err))
			return
		}
		evt = event.NewEvent(event.SyncDesireEvent, pld)
	} else {
		pld, err := json.Marshal(info)
		if err != nil {
			e.log.Error("failed to marshal delta", log.Error(err))
			return
		}
		evt = event.NewEvent(event.SyncReportEvent, pld)
	}
	err = e.cent.Trigger(evt)
	if err != nil {
		e.log.Error("failed to trigger event", log.Error(err))
		return
	}
}

func (e *Engine) Close() {
	e.tomb.Kill(nil)
	e.tomb.Wait()
}

func (e *Engine) Apply(evt *event.Event) error {
	var info v1.Desire
	err := json.Unmarshal(evt.Payload, &info)
	if err != nil {
		return err
	}
	apps := info.AppInfos()
	if len(apps) == 0 {
		return fmt.Errorf("apps does not exist")
	}
	err = e.ami.Apply(apps)
	if err != nil {
		e.log.Error("failed to apply application", log.Error(err))
		return err
	}
	return nil
}
