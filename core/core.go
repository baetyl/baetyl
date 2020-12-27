package core

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	"github.com/valyala/fasthttp"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/engine"
	"github.com/baetyl/baetyl/v2/eventx"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/store"
	"github.com/baetyl/baetyl/v2/sync"
	"github.com/baetyl/baetyl/v2/utils"
)

type NewCoreFunc func(ctx context.Context, cfg config.Config) (Core, error)

type Core interface {
	Close()
}

type core struct {
	cfg config.Config
	sto *bh.Store
	nod node.Node
	eng engine.Engine
	syn sync.Sync
	svr *http.Server
	evt eventx.EventX
}

// NewCore creates a new core
func NewCore(ctx context.Context, cfg config.Config) (Core, error) {
	err := utils.ExtractNodeInfo(cfg.Node)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := &core{}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.nod, err = node.NewNode(c.sto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.syn, err = sync.NewSync(cfg, c.sto, c.nod)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.eng, err = engine.NewEngine(cfg, c.sto, c.nod, c.syn)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.svr = http.NewServer(cfg.Server, c.initRouter())

	c.eng.Start()
	c.syn.Start()
	c.svr.Start()

	if cfg.Event.Notify {
		c.evt, err = eventx.NewEventX(ctx, cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		c.evt.Start()
	}
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
	if c.evt != nil {
		c.evt.Close()
	}
}

func (c *core) initRouter() fasthttp.RequestHandler {
	router := routing.New()
	router.Get("/node/stats", utils.Wrapper(c.nod.GetStats))
	router.Get("/services/<service>/log", c.eng.GetServiceLog)
	router.Get("/node/properties", utils.Wrapper(c.nod.GetNodeProperties))
	router.Put("/node/properties", utils.Wrapper(c.nod.UpdateNodeProperties))
	return router.HandleRequest
}
