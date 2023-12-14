package initz

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	v2utils "github.com/baetyl/baetyl-go/v2/utils"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/ami/kube"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/engine"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/baetyl/baetyl/v2/store"
	"github.com/baetyl/baetyl/v2/sync"
	"github.com/baetyl/baetyl/v2/utils"
)

type StartInitServiceFunc func()

type Initialize struct {
	cfg  config.Config
	sto  *bh.Store
	nod  node.Node
	eng  engine.Engine
	syn  sync.Sync
	log  *log.Logger
	tomb v2utils.Tomb
}

// NewInitialize creates a new core
func NewInitialize(cfg config.Config) (*Initialize, error) {
	if err := initHooks(cfg); err != nil {
		return nil, errors.Trace(err)
	}
	// to activate if no node cert
	if !v2utils.FileExists(cfg.Node.Cert) {
		active, err := NewActivate(&cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		active.Start()
		active.WaitAndClose()
		log.L().Info("init activates node success")
	}

	err := utils.ExtractNodeInfo(cfg.Node)
	if err != nil {
		return nil, errors.Trace(err)
	}

	init := &Initialize{
		cfg: cfg,
		log: log.With(log.Any("init", "sync")),
	}
	init.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}

	init.nod, err = node.NewNode(init.sto)
	if err != nil {
		return nil, errors.Trace(err)
	}

	init.syn, err = sync.NewSync(cfg, init.sto, init.nod)
	if err != nil {
		return nil, errors.Trace(err)
	}

	init.eng, err = engine.NewEngine(cfg, init.sto, init.nod, init.syn, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	init.eng.Start()
	init.syn.Start()
	return init, nil
}

func initHooks(cfg config.Config) error {
	ami.Hooks[kube.BaetylSetPodSpec] = kube.SetPodSpecFunc(kube.SetPodSpec)
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
	return nil
}

func (init *Initialize) Close() {
	if init.eng != nil {
		init.eng.Close()
	}
	if init.sto != nil {
		init.sto.Close()
	}
	if init.syn != nil {
		init.syn.Close()
	}
}

func StartInitService() {
	context.Run(func(ctx context.Context) error {
		var cfg config.Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return errors.Trace(err)
		}
		plugin.ConfFile = ctx.ConfFile()

		init, err := NewInitialize(cfg)
		if err != nil {
			return errors.Trace(err)
		}
		defer init.Close()

		ctx.Wait()
		return nil
	})
}
