package main

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/omi"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/log"
	bh "github.com/timshannon/bolthold"
)

type core struct {
	s       sync.Sync
	store   *bh.Store
	cfg     config.Config
	engine  *engine.Engine
}

func NewCore(ctx context.Context, cfg config.Config) (*core, error) {
	logger, err := log.Init(cfg.Logger)
	if err != nil {
		return nil, err
	}
	store, err := store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}
	model, err := omi.NewKubeModel(cfg.APIServer, store)
	if err != nil {
		return nil, err
	}
	e := engine.NewEngine(cfg.Interval, model, logger)
	s, err := sync.NewSync(ctx, cfg.Sync, store, logger)
	if err != nil {
		return nil, err
	}
	return &core{
		engine: e,
		store:  store,
		cfg:    cfg,
		s:      s,
	}, nil
}

func (c *core) Stop() {
	c.engine.Close()
	c.s.Stop()
	c.store.Close()
}

func main() {
	context.Run(func(ctx context.Context) error {
		var cfg config.Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return err
		}
		c, err := NewCore(ctx, cfg)
		if err != nil {
			return err
		}
		defer c.Stop()
		ctx.Wait()
		return nil
	})
}
