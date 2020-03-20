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
	"github.com/baetyl/baetyl-go/faas"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/api"
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
	e.tomb.Go(e.collecting)
	return e, nil
}

func (e *Engine) collecting() error {
	t := time.NewTicker(e.cfg.Collector.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			info, err := e.ami.CollectInfo()
			if err != nil {
				e.log.Error("failed to collect info", log.Error(err))
				continue
			}
			// TODO: improve
			rep := map[string]interface{}{
				"time":      info.Time,
				"node":      info.NodeInfo,
				"nodestats": info.NodeStat,
				"apps":      info.AppInfos,
				"appstats":  info.AppStats,
			}
			delta, err := e.sha.Report(rep)
			if err != nil {
				e.log.Error("failed to update shadow report", log.Error(err))
				continue
			}
			var msg faas.Message
			if len(delta) > 0 {
				msg.Payload, err = json.Marshal(delta)
				if err != nil {
					e.log.Error("failed to marshal delta", log.Error(err))
					continue
				}
				msg.Metadata = map[string]string{"topic": event.SyncDesireEvent}
			} else {
				msg.Payload, err = json.Marshal(info)
				if err != nil {
					e.log.Error("failed to marshal delta", log.Error(err))
					continue
				}
				msg.Metadata = map[string]string{"topic": event.SyncReportEvent}
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

func (e *Engine) Apply(msg faas.Message) error {
	var info api.ReportResponse
	err := json.Unmarshal(msg.Payload, &info)
	if err != nil {
		return err
	}
	if len(info.AppInfos) == 0 {
		return fmt.Errorf("apps does not exist")
	}
	err = e.ami.ApplyApplications(&info)
	if err != nil {
		return err
	}
	return nil
}
