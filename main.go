package main

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/init"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	bh "github.com/timshannon/bolthold"
)

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *shadow.Shadow
	eng *engine.Engine
	syn sync.Sync
}

func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Node.Name == "" {
		i, err := init.NewInit(cfg)
		if err != nil {
			i.Close()
			return nil, err
		}

	}

	c := &core{}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}
	c.sha, err = shadow.NewShadow(cfg.Node.Namespace, cfg.Node.Name, c.sto)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.eng, err = engine.NewEngine(cfg.Engine, c.sto, c.sha)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *core) Close() {
	if c.syn != nil {
		c.syn.Close()
	}
	if c.eng != nil {
		c.eng.Close()
	}
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
		defer c.Close()
		ctx.Wait()
		return nil
	})
}
