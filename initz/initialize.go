package initz

import (
	"fmt"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/engine"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/store"
	"github.com/baetyl/baetyl/v2/sync"
)

var (
	// ErrSysAppCoreMissing system application baetyl-core is required for connection with cloud
	ErrSysAppCoreMissing = fmt.Errorf("system application baetyl-core is required for connection with cloud")
)

const (
	BaetylCore = "baetyl-core"
)

type Initialize struct {
	cfg  config.Config
	sto  *bh.Store
	sha  *node.Node
	eng  engine.Engine
	syn  sync.Sync
	log  *log.Logger
	tomb utils.Tomb
}

// NewInitialize creates a new core
func NewInitialize(cfg config.Config) (*Initialize, error) {
	// to activate if no node cert
	if !utils.FileExists(cfg.Node.Cert) {
		active, err := NewActivate(&cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		active.Start()
		active.WaitAndClose()
		log.L().Info("init activates node success")
	}

	var err error
	init := &Initialize{
		cfg: cfg,
		log: log.With(log.Any("init", "sync")),
	}
	init.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}

	init.sha, err = node.NewNode(init.sto)
	if err != nil {
		init.Close()
		return nil, errors.Trace(err)
	}

	init.syn, err = sync.NewSync(cfg, init.sto, init.sha)
	if err != nil {
		init.Close()
		return nil, errors.Trace(err)
	}

	init.eng, err = engine.NewEngine(cfg, init.sto, init.sha, init.syn)
	if err != nil {
		init.Close()
		return nil, errors.Trace(err)
	}
	init.tomb.Go(init.start)
	return init, nil
}

func (init *Initialize) start() error {
	defer init.closeStore()
	init.log.Info("collect and report stats to cloud")
	err := init.reportAndDesireCloud()
	if err != nil {
		return errors.Trace(err)
	}

	init.log.Info("collect and report stats, then apply applications at edge")
	err = init.eng.ReportAndDesire()
	if err != nil {
		return errors.Trace(err)
	}

	// close store which shared with baetyl-core
	init.closeStore()

	t := time.NewTicker(init.cfg.Sync.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			init.log.Info("collect stats from edge", log.Error(err))
			r := init.eng.Collect("baetyl-edge-system", true, nil)
			init.log.Info("report stats to cloud", log.Error(err))
			_, err = init.syn.Report(r)
			if err != nil {
				return errors.Trace(err)
			}
		case <-init.tomb.Dying():
			return nil
		}
	}
}

func (init *Initialize) closeStore() {
	if init.sto != nil {
		err := init.sto.Close()
		if err != nil {
			init.log.Warn("failed to close store", log.Error(err))
		} else {
			init.log.Info("close store")
		}
		init.sto = nil
	}
}

func (init *Initialize) Close() {
	init.tomb.Kill(nil)
	init.tomb.Wait()

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

func (init *Initialize) reportAndDesireCloud() error {
	r := init.eng.Collect("baetyl-edge-system", true, nil)
	ds, err := init.syn.Report(r)
	if err != nil {
		init.log.Error("failed to report app info", log.Error(err))
		return errors.Trace(ErrSysAppCoreMissing)
	}
	init.log.Debug("cloud response shadow desire", log.Any("desire", ds))
	if len(ds) == 0 {
		return errors.Trace(ErrSysAppCoreMissing)
	}

	for _, app := range ds.AppInfos(true) {
		if strings.Contains(app.Name, BaetylCore) {
			n := specv1.Desire{}
			n.SetAppInfos(true, []specv1.AppInfo{app})
			_, err = init.sha.Desire(n)
			return errors.Trace(err)
		}
	}
	return errors.Trace(ErrSysAppCoreMissing)
}
