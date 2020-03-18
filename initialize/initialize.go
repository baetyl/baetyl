package initialize

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

//go:generate mockgen -destination=../mock/initialize.go -package=mock github.com/baetyl/baetyl-core/initialize Initialize

type batch struct {
	name         string
	namespace    string
	securityType string
	securityKey  string
}

type Initialize interface {
	Start()
	Close()
	Activate()
	WaitAndClose()
}

// NewInit to activate, success add node info
func NewInit(cfg *config.Config) (Initialize, error) {
	httpOps, err := cfg.Init.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}
	httpCli := http.NewClient(*httpOps)
	init := &initialize{
		cfg:   cfg,
		http:  httpCli,
		sig:   make(chan bool, 1),
		attrs: map[string]string{},
		log:   log.With(log.Any("core", "initialize")),
	}
	init.batch = &batch{
		name:         cfg.Init.Batch.Name,
		namespace:    cfg.Init.Batch.Namespace,
		securityType: cfg.Init.Batch.SecurityType,
		securityKey:  cfg.Init.Batch.SecurityKey,
	}
	for _, a := range cfg.Init.ActivateConfig.Attributes {
		init.attrs[a.Name] = a.Value
	}
	init.Start()
	return init, nil
}

type initialize struct {
	log   *log.Logger
	cfg   *config.Config
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
		init.log.Error("initialize get sig error")
	}
	init.Close()
}
