package engine

import (
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

type Engine struct {
	sha  *node.Node
	cfg  config.EngineConfig
	ami  ami.AMI
	tomb utils.Tomb
	log  *log.Logger
}

func NewEngine(cfg config.EngineConfig, sto *bh.Store, sha *node.Node) (*Engine, error) {
	if cfg.Kind != "kubernetes" {
		return nil, os.ErrInvalid
	}
	ami, err := ami.NewKubeImpl(cfg.Kubernetes, sto)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		sha: sha,
		ami: ami,
		cfg: cfg,
		log: log.With(log.Any("engine", cfg.Kind)),
	}
	e.tomb.Go(e.reporting)
	return e, nil
}

func (e *Engine) reporting() error {
	e.log.Info("engine starts to report")
	defer e.log.Info("engine has stopped reporting")

	t := time.NewTicker(e.cfg.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := e.report()
			if err != nil {
				e.log.Error("failed to report local shadow", log.Error(err))
			} else {
				e.log.Debug("engine reports local shadow")
			}
		case <-e.tomb.Dying():
			return nil
		}
	}
}

func (e *Engine) report() error {
	// to collect app status
	info, err := e.ami.Collect()
	if err != nil {
		return err
	}
	if len(info) == 0 {
		return errors.New("no status collected")
	}
	// to report app status into local shadow, and return shadow delta
	delta, err := e.sha.Report(info)
	if err != nil {
		return err
	}
	// if apps are updated, to apply new apps
	if delta == nil {
		return nil
	}
	apps := delta.AppInfos()
	if apps == nil {
		return nil
	}
	e.log.Info("to apply apps", log.Any("apps", apps))
	return e.ami.Apply(apps)
}

func (e *Engine) Close() {
	e.tomb.Kill(nil)
	e.tomb.Wait()
}
