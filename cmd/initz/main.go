package main

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/initz"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

var (
	// ErrSysAppCoreMissing system application baetyl-core is required for connection with cloud
	ErrSysAppCoreMissing = fmt.Errorf("system application baetyl-core is required for connection with cloud")
)

const (
	BaetylCore = "baetyl-core"
)

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn sync.Sync
	log *log.Logger
}

// NewCore creats a new core
func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c := &core{
		cfg: cfg,
		log: log.With(log.Any("initz", "main")),
	}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.sha, err = node.NewNode(c.sto)
	if err != nil {
		c.Close()
		return nil, errors.Trace(err)
	}

	if !utils.FileExists(cfg.Sync.Cloud.HTTP.Cert) {
		i, err := initz.NewInit(&cfg)
		if err != nil {
			i.Close()
			return nil, errors.Trace(err)
		}
		i.Start()
		i.WaitAndClose()
		c.log.Info("init active success")
	}

	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.eng, err = engine.NewEngine(cfg.Engine, c.sto, c.sha, c.syn)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return c, nil
}

func (c *core) Close() {
	if c.eng != nil {
		c.eng.Close()
	}
	if c.sto != nil {
		c.sto.Close()
	}
	if c.syn != nil {
		c.syn.Close()
	}
}

func (c *core) reportAndDesireCloud() error {
	r, err := c.eng.Collect("baetyl-edge-system", true)
	if err != nil {
		return errors.Trace(err)
	}
	ds, err := c.syn.Report(r)
	if err != nil {
		c.log.Error("failed to report app info", log.Error(err))
		return errors.Trace(ErrSysAppCoreMissing)
	}
	if len(ds) == 0 {
		return errors.Trace(ErrSysAppCoreMissing)
	}

	for _, app := range ds.AppInfos(true) {
		if strings.Contains(app.Name, BaetylCore) {
			n := specv1.Desire{}
			n.SetAppInfos(true, []specv1.AppInfo{app})
			_, err = c.sha.Desire(n)
			return errors.Trace(err)
		}
	}
	return errors.Trace(ErrSysAppCoreMissing)
}

func main() {
	context.Run(func(ctx context.Context) error {
		c, err := NewCore(ctx)
		if err != nil {
			return errors.Trace(err)
		}
		defer c.Close()

		err = c.reportAndDesireCloud()
		if err != nil {
			return errors.Trace(err)
		}
		err = c.eng.ReportAndDesire()
		if err != nil {
			return errors.Trace(err)
		}
		if c.sto != nil {
			c.sto.Close()
			c.sto = nil
		}

		t := time.NewTicker(c.cfg.Sync.Cloud.Report.Interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				c.log.Info("collect stats from edge", log.Error(err))
				r, err := c.eng.Collect("baetyl-edge-system", true)
				if err != nil {
					c.log.Error("failed to collect info", log.Error(err))
				}
				c.log.Info("report stats to cloud", log.Error(err))
				_, err = c.syn.Report(r)
				if err != nil {
					c.log.Error("failed to report info", log.Error(err))
				}
			case <-ctx.WaitChan():
				return nil
			}
		}
	})
}
