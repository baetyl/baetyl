package main

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/engine"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-core/sync"
	"github.com/baetyl/baetyl-go/context"
	"github.com/baetyl/baetyl-go/http"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	"github.com/valyala/fasthttp"
)

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn *sync.Sync
	svr *http.Server
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
	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.syn.Start()

	c.eng, err = engine.NewEngine(cfg.Engine, c.sto, c.sha, c.syn)
	if err != nil {
		c.Close()
		return nil, err
	}

	c.eng.Start()

	c.svr = http.NewServer(cfg.Server, c.initRouter())
	c.svr.Start()
	return c, nil
}

func (c *core) Close() {
	if c.svr != nil {
		c.svr.Close()
	}
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

func (c *core) initRouter() fasthttp.RequestHandler {
	router := routing.New()
	router.Get("/node/status", c.sha.GetStatus)
	router.Get("/services/<service>/log", c.eng.GetServiceLog)
	return router.HandleRequest
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
