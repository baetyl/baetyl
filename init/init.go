package init

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
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
	WaitAndClose()
}

func NewInit(cfg config.Config) (Init, error) {
	httpOps, err := cfg.Init.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}
	httpCli := http.NewClient(*httpOps)
	init := &initialize{
		cfg:  cfg,
		http: httpCli,
		sig:  make(chan bool, 1),
		log:  log.With(log.Any("core", "init")),
	}
	init.batch = &batch{
		Name:         cfg.Init.Batch.Name,
		Namespace:    cfg.Init.Batch.Namespace,
		SecurityType: cfg.Init.Batch.SecurityType,
		SecurityKey:  cfg.Init.Batch.SecurityKey,
	}
	return init, nil
}

type initialize struct {
	log   *log.Logger
	cfg   config.Config
	tomb  utils.Tomb
	http  *http.Client
	batch *batch
	attrs map[string]string
	sig   chan bool
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

func (init *initialize) WaitAndClose() {
	if _, ok := <-init.sig; !ok {
		init.log.Error("init get sig error")
	}
	init.Close()
}
