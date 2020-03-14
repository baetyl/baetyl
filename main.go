package main

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/kube"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/log"
	bh "github.com/timshannon/bolthold"
)

type core struct {
	s       sync.Sync
	kubeCli *kube.Client
	store   *bh.Store
	cfg     config.Config
}

func NewCore(ctx context.Context, cfg config.Config) (*core, error) {
	kubeCli, err := kube.NewClient(cfg.APIServer)
	logger, err := log.Init(cfg.Logger)
	if err != nil {
		return nil, err
	}
	store, err := store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}
	s, err := sync.NewSync(ctx, cfg.Sync, kubeCli.AppV1.Deployments(kubeCli.Namespace), store, logger)
	if err != nil {
		return nil, err
	}
	return &core{
		kubeCli: kubeCli,
		store:   store,
		cfg:     cfg,
		s:       s,
	}, nil
}

func (c *core) Start() error {
	go c.s.Start()
	return nil
}

func (c *core) Stop() {
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
		err = c.Start()
		if err != nil {
			return err
		}
		ctx.Wait()
		return nil
	})
}
