package main

import (
	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/event"
	"github.com/baetyl/baetyl-core/initialize"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	bh "github.com/timshannon/bolthold"
)

type core struct {
	cfg  config.Config
	sto  *bh.Store
	cent *event.Center
	sha  *shadow.Shadow
	eng  *engine.Engine
	syn  *sync.Sync
}

func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Sync.Node.Name == "" {
		i, err := initialize.NewInit(&cfg)
		if err != nil {
			i.Close()
			return nil, err
		}
		i.WaitAndClose()
	}

	c := &core{}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}
	model, err := ami.NewKubeModel(cfg.Engine.Kubernetes, c.sto)
	if err != nil {
		return nil, err
	}
	c.cent, err = event.NewCenter(c.sto, common.EventCenterLimit)
	if err != nil {
		return nil, err
	}
	c.sha, err = shadow.NewShadow(cfg.Sync.Node.Namespace, cfg.Sync.Node.Name, c.sto)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.eng, err = engine.NewEngine(cfg.Engine, model, c.sha, c.cent)
	if err != nil {
		c.Close()
		return nil, err
	}
	err = c.cent.Register(common.EngineAppEvent, c.eng.Apply)
	if err != nil {
		return nil, err
	}

	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha, c.cent)
	if err != nil {
		c.Close()
		return nil, err
	}
	err = c.cent.Register(common.SyncReportEvent, c.syn.Report)
	if err != nil {
		c.Close()
		return nil, err
	}
	err = c.cent.Register(common.SyncDesireEvent, c.syn.ProcessDelta)
	if err != nil {
		c.Close()
		return nil, err
	}

	c.cent.Start()
	return c, nil
}

func (c *core) Close() {
	if c.eng != nil {
		c.eng.Close()
	}
	if c.sto != nil {
		c.sto.Close()
	}
	if c.cent != nil {
		c.cent.Close()
	}
}

func main() {
	context.Run(func(ctx context.Context) error {
		c, err := NewCore(ctx)
		if err != nil {
			return err
		}
		defer c.Close()
		ctx.Wait()
		return nil
	})
}
