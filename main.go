package main

import (
	"encoding/json"
	"time"

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

const OfflineDuration = 40 * time.Second

type core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn *sync.Sync
	svc *http.Server
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
		c.Close()
		return nil, err
	}
	c.eng.Start()
	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		c.Close()
		return nil, err
	}
	c.syn.Start()
	c.svc = http.NewServer(cfg.Server, c.initRouter())
	c.svc.Start()
	return c, nil
}

func (c *core) Close() {
	if c.svc != nil {
		c.svc.Close()
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
	router.Get("/shadow", c.getStatus)
	return router.HandleRequest
}

func (c *core) getStatus(ctx *routing.Context) error {
	node, err := c.sha.Get()
	if err != nil {
		http.RespondMsg(ctx, 500, "ERR_DB", err.Error())
		return nil
	}

	view := node.View(OfflineDuration)
	res, err := json.Marshal(view)
	if err != nil {
		http.RespondMsg(ctx, 500, "ERR_JSON", err.Error())
		return nil
	}
	http.Respond(ctx, 200, res)
	return nil
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
