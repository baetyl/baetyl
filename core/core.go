package core

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	"github.com/valyala/fasthttp"

	"github.com/baetyl/baetyl/v2/agent"
	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/ami/kube"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/dm"
	"github.com/baetyl/baetyl/v2/engine"
	"github.com/baetyl/baetyl/v2/eventx"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/baetyl/baetyl/v2/roam"
	"github.com/baetyl/baetyl/v2/store"
	"github.com/baetyl/baetyl/v2/sync"
	"github.com/baetyl/baetyl/v2/utils"
)

type StartCoreServiceFunc func()

type Core struct {
	cfg config.Config
	sto *bh.Store
	nod node.Node
	eng engine.Engine
	syn sync.Sync
	svr *http.Server
	agt agent.AgentClient
	evt eventx.EventX
	dm  dm.DeviceManager
}

// NewCore creates a new core
func NewCore(ctx context.Context, cfg config.Config) (*Core, error) {
	err := utils.ExtractNodeInfo(cfg.Node)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := &Core{}
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
	c.agt, err = agent.NewAgentClient(c.syn)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.eng, err = engine.NewEngine(cfg, c.sto, c.nod, c.syn, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if err = initHooks(cfg); err != nil {
		return nil, errors.Trace(err)
	}

	c.svr = http.NewServer(cfg.Server, c.initRouter())
	c.dm, err = dm.NewDeviceManager(ctx, c.sto, c.nod, c.syn, cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.eng.Start()
	c.syn.Start()
	c.svr.Start()
	c.dm.Start()

	if cfg.Event.Notify {
		c.evt, err = eventx.NewEventX(ctx, cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		c.evt.Start()
	}
	return c, nil
}

func initHooks(cfg config.Config) error {
	// set GPU hook function
	if cfg.StatsExt.GPU {
		pl, err := v2plugin.GetPlugin(cfg.ExtPlugin.GpuStats)
		if err != nil {
			return errors.Trace(err)
		}
		stats := pl.(plugin.Collect)
		ami.Hooks[ami.BaetylGPUStatsExtension] = ami.CollectStatsExtFunc(stats.CollectStats)
		log.L().Info("registered gpu stats collector")
	}
	// set Node Disk&Net Stats hook function
	if cfg.StatsExt.NodeStats {
		nodePlugin, err := v2plugin.GetPlugin(cfg.ExtPlugin.NodeStats)
		if err != nil {
			return errors.Trace(err)
		}
		nodeStats := nodePlugin.(plugin.Collect)
		ami.Hooks[ami.BaetylNodeStatsExtension] = ami.CollectStatsExtFunc(nodeStats.CollectStats)
		log.L().Info("registered node stats collector")
	}
	if cfg.StatsExt.QPSStats {
		qpsPlugin, err := v2plugin.GetPlugin(cfg.ExtPlugin.QPSStats)
		if err != nil {
			return errors.Trace(err)
		}
		qpsStats := qpsPlugin.(plugin.Collect)
		ami.Hooks[ami.BaetylQPSStatsExtension] = ami.CollectStatsExtFunc(qpsStats.CollectStats)
		log.L().Info("registered qps stats collector")
	}

	// set ami prepare deployment function
	ami.Hooks[kube.BaetylSetPodSpec] = kube.SetPodSpecFunc(kube.SetPodSpec)

	// set sync upload object function
	r, err := roam.NewRoam(cfg)
	if err != nil {
		return errors.Trace(err)
	}
	sync.Hooks[sync.BaetylHookUploadObject] = sync.UploadObjectFunc(r.RoamObject)
	return nil
}

func (c *Core) Close() {
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
	if c.dm != nil {
		c.dm.Close()
	}
}

func (c *Core) initRouter() fasthttp.RequestHandler {
	router := routing.New()
	router.Get("/node/stats", utils.Wrapper(c.nod.GetStats))
	router.Get("/services/<service>/log", c.eng.GetServiceLog)
	router.Get("/node/properties", utils.Wrapper(c.nod.GetNodeProperties))
	router.Put("/node/properties", utils.Wrapper(c.nod.UpdateNodeProperties))
	router.Post("/agent/sts", utils.Wrapper(c.agt.SendRequest))
	router.Get("/sync/state", utils.Wrapper(c.syn.LinkState))
	return router.HandleRequest
}

func StartCoreService() {
	context.Run(func(ctx context.Context) error {
		var cfg config.Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return errors.Trace(err)
		}
		plugin.ConfFile = ctx.ConfFile()

		c, err := NewCore(ctx, cfg)
		if err != nil {
			return errors.Trace(err)
		}
		defer c.Close()

		ctx.Wait()
		return nil
	})
}
