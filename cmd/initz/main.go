package main

import (
	"errors"
	"strings"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/initialize"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/log"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

var (
	// ErrSysappCoreMissing system application baetyl-core is required for connection with cloud
	ErrSysappCoreMissing = errors.New("system application baetyl-core is required for connection with cloud")
)

const (
	BaetylCore = "baetyl-core"
)

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn *sync.Sync
	log *log.Logger
}

// NewCore creats a new core
func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, err
	}

	c := &core{
		cfg: cfg,
		log: log.With(log.Any("initz", "main")),
	}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}

	c.sha, err = node.NewNode(c.sto)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.eng, err = engine.NewEngine(cfg.Engine, c.sto, c.sha)
	if err != nil {
		return nil, err
	}

	if !utils.FileExists(cfg.Sync.Cloud.HTTP.Cert) {
		i, err := initialize.NewInit(&cfg, c.eng.Ami)
		if err != nil {
			i.Close()
			return nil, err
		}
		i.Start()
		i.WaitAndClose()
	}

	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *core) Close() {
	if c.sto != nil {
		c.sto.Close()
	}
}

func (c *core) reportAndDesire() error {
	r, err := c.eng.Ami.Collect("baetyl-edge")
	if err != nil {
		return err
	}
	ds, err := c.syn.Report(r)
	c.log.Debug("init report info", log.Any("report", r))
	c.log.Debug("init desire info", log.Any("desire", ds))
	if err != nil {
		c.log.Error("sync report error", log.Any("sync", err))
		return ErrSysappCoreMissing
	}
	if len(ds) == 0 || len(ds.SysAppInfos()) == 0 {
		return ErrSysappCoreMissing
	}

	for _, app := range ds.SysAppInfos() {
		if strings.Contains(app.Name, BaetylCore) {
			err := c.syn.Desire(v1.Desire{
				"sysapps": []v1.AppInfo{app},
			})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return ErrSysappCoreMissing
}

func main() {
	context.Run(func(ctx context.Context) error {
		c, err := NewCore(ctx)
		if err != nil {
			return err
		}
		defer c.Close()

		for {
			err := c.reportAndDesire()
			if err == nil {
				break
			}
			if err != ErrSysappCoreMissing {
				c.log.Error("init get core error", log.Any("error", err))
				return err
			}
			time.Sleep(c.cfg.Sync.Cloud.Report.Interval)
		}
		err = c.eng.ReportAndDesire()
		if err != nil {
			return err
		}
		err = c.sto.Close()
		if err != nil {
			return err
		}

		t := time.NewTicker(c.cfg.Sync.Cloud.Report.Interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				r, err := c.eng.Ami.Collect("baetyl-edge")
				if err != nil {
					c.log.Error("init collect error", log.Any("collect", err))
				}
				if _, err := c.syn.Report(r); err != nil {
					c.log.Error("init report error", log.Any("report", err))
				}
			case <-ctx.WaitChan():
				return nil
			}
		}
	})
}
