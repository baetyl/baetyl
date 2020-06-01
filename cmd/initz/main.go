package main

import (
	"errors"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"strings"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/initz"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/log"
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

	if !utils.FileExists(cfg.Sync.Cloud.HTTP.Cert) {
		i, err := initz.NewInit(&cfg, c.eng.Ami)
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

	c.eng, err = engine.NewEngine(cfg.Engine, c.sto, c.sha, c.syn)
	if err != nil {
		return nil, err
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

func (c *core) reportAndDesire() error {
	r, err := c.eng.Collect("baetyl-edge-system", true)
	if err != nil {
		return err
	}
	ds, err := c.syn.Report(r)
	if err != nil {
		c.log.Error("failed to report app info", log.Error(err))
		return ErrSysappCoreMissing
	}
	if len(ds) == 0 {
		return ErrSysappCoreMissing
	}

	for _, app := range ds.AppInfos(true) {
		if strings.Contains(app.Name, BaetylCore) {
			if _, err := c.sha.Desire(specv1.Desire{"sysapps": []specv1.AppInfo{app}}); err != nil {
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

		err = c.reportAndDesire()
		if err != nil {
			return err
		}
		err = c.eng.ReportAndDesire()
		if err != nil {
			return err
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
				r, err := c.eng.Collect("baetyl-edge-system", true)
				if err != nil {
					c.log.Error("failed to collect info", log.Error(err))
				}
				if _, err := c.syn.Report(r); err != nil {
					c.log.Error("failed to report info", log.Error(err))
				}
			case <-ctx.WaitChan():
				return nil
			}
		}
	})
}
