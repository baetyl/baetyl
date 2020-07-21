package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	_ "github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/engine"
	"github.com/baetyl/baetyl/initz"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/baetyl/baetyl/sync"
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

// NewCore creates a new core
func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// to activate if no node cert
	if !utils.FileExists(cfg.Sync.Cloud.HTTP.Cert) {
		i, err := initz.NewInit(&cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		i.Start()
		i.WaitAndClose()
		ctx.Log().Info("init activates node success")
	}

	c := &core{
		cfg: cfg,
		log: log.With(log.Any("init", "sync")),
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

	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		c.Close()
		return nil, errors.Trace(err)
	}

	c.eng, err = engine.NewEngine(cfg, c.sto, c.sha, c.syn)
	if err != nil {
		c.Close()
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
	r := c.eng.Collect("baetyl-edge-system", true, nil)
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

		c.log.Info("collect and report stats to cloud")
		err = c.reportAndDesireCloud()
		if err != nil {
			return errors.Trace(err)
		}

		c.log.Info("collect and report stats, then apply applications at edge")
		err = c.eng.ReportAndDesire()
		if err != nil {
			return errors.Trace(err)
		}

		// close store which shared with baetyl-core
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
				r := c.eng.Collect("baetyl-edge-system", true, nil)
				c.log.Info("report stats to cloud", log.Error(err))
				_, err = c.syn.Report(r)
				if err != nil {
					return errors.Trace(err)
				}
			case <-ctx.WaitChan():
				return nil
			}
		}
	})
}
