package main

import (
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/initialize"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
)

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn *sync.Sync
}

// NewCore creats a new core
func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, err
	}

	c := &core{}
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

func main() {
	context.Run(func(ctx context.Context) error {
		c, err := NewCore(ctx)
		if err != nil {
			return err
		}

		err = c.syn.ReportAndDesire()
		if err != nil {
			return err
		}
		err = c.eng.ReportAndDesire()
		if err != nil {
			return err
		}
		c.Close()

		for {
			r, err := c.eng.Ami.Collect("baetyl-edge")
			if err != nil {
				ctx.Log().Error("init collect error", log.Any("Main", err))
			}
			if err := c.syn.Report(r); err != nil {
				ctx.Log().Error("init report error", log.Any("Main", err))
			}
			time.Sleep(c.cfg.Sync.Cloud.Report.Interval)
		}
	})
}
