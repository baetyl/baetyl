package main

import (
	"encoding/json"
	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/event"
	"github.com/baetyl/baetyl-core/initialize"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
	"os"
	"path"
	"runtime"
)

type core struct {
	cfg  config.Config
	sto  *bh.Store
	cent *event.Center
	sha  *shadow.Shadow
	eng  *engine.Engine
	syn  *sync.Sync
}

// NewCore creats a new core
func NewCore(ctx context.Context) (*core, error) {
	var cfg config.Config
	err := ctx.LoadCustomConfig(&cfg)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(path.Dir(cfg.Store.Path), 0755)
	if err != nil {
		return nil, err
	}
	c := &core{}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, err
	}
	ami, err := ami.NewKubeImpl(cfg.Engine.Kubernetes, c.sto)
	if err != nil {
		return nil, err
	}
	if cfg.Sync.Node.Name == "" {
		i, err := initialize.NewInit(&cfg, ami)
		if err != nil {
			i.Close()
			return nil, err
		}
		i.WaitAndClose()
	}
	c.cent, err = event.NewCenter(c.sto, 10)
	if err != nil {
		return nil, err
	}
	c.sha, err = shadow.NewShadow(cfg.Sync.Node.Namespace, cfg.Sync.Node.Name, c.sto)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.eng, err = engine.NewEngine(cfg.Engine, ami, c.sha, c.cent)
	if err != nil {
		c.Close()
		return nil, err
	}
	err = c.cent.Register(event.EngineAppEvent, c.eng.Apply)
	if err != nil {
		return nil, err
	}

	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha, c.cent)
	if err != nil {
		c.Close()
		return nil, err
	}
	err = c.cent.Register(event.SyncReportEvent, c.syn.Report)
	if err != nil {
		c.Close()
		return nil, err
	}
	err = c.cent.Register(event.SyncDesireEvent, c.syn.Desire)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.eng.Start()
	c.cent.Start()
	r := v1.Report{
		"core": v1.CoreInfo{
			GoVersion:   runtime.Version(),
			BinVersion:  utils.VERSION,
			GitRevision: utils.REVISION,
		},
	}
	_, err = c.sha.Report(r)
	if err != nil {
		return nil, err
	}
	pld, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	evt := event.NewEvent(event.SyncReportEvent, pld)
	err = c.cent.Trigger(evt)
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
