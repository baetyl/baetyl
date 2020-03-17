package engine

import (
	"encoding/json"
	"os"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/omi"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

type Engine struct {
	cfg   config.EngineConfig
	model omi.Model
	tomb  utils.Tomb
	log   *log.Logger
}

func NewEngine(cfg config.EngineConfig, sto *bh.Store, sha *shadow.Shadow) (*Engine, error) {
	if cfg.Kind != "kubernetes" {
		return nil, os.ErrInvalid
	}
	e := &Engine{
		cfg: cfg,
		log: log.With(log.Any("engine", cfg.Kind)),
	}
	var err error
	e.model, err = omi.NewKubeModel(cfg.Kubernetes, sto, sha)
	if err != nil {
		return nil, err
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
