package init

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	bh "github.com/timshannon/bolthold"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

//go:generate mockgen -destination=../mock/init.go -package=mock github.com/baetyl/baetyl-core/init Init

type batch struct {
	Name         string
	Namespace    string
	SecurityType string
	SecurityKey  string
}

type Init interface {
	Start()
	Close()
	Activate()
}

func NewInit(cfg config.Config, sto *bh.Store, sha *shadow.Shadow) (Init, error) {
	httpOps, err := cfg.Init.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}
	httpCli := http.NewClient(*httpOps)
	init := &initialize{
		cfg:    cfg,
		store:  sto,
		shadow: sha,
		events: make(chan *Event, 1),
		http:   httpCli,
		log:    log.With(log.Any("core", "init")),
	}
	init.batch = &batch{
		Name:         cfg.Batch.Name,
		Namespace:    cfg.Batch.Namespace,
		SecurityType: cfg.Batch.SecurityType,
		SecurityKey:  cfg.Batch.SecurityKey,
	}
	return init, nil
}

type initialize struct {
	log    *log.Logger
	cfg    config.Config
	tomb   utils.Tomb
	impl   appv1.DeploymentInterface
	events chan *Event
	http   *http.Client
	batch  *batch
	store  *bh.Store
	shadow *shadow.Shadow
}

func (init *initialize) Start() {
	err := init.tomb.Go(init.activating)
	if err != nil {
		init.log.Error("failed to start report and process routine", log.Error(err))
		return
	}
}

func (init *initialize) Close() {
	init.tomb.Kill(nil)
	init.tomb.Wait()
}
