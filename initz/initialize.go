package initz

import (
	gohttp "net/http"

	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

type batch struct {
	name         string
	namespace    string
	securityType string
	securityKey  string
}

type Initialize struct {
	log   *log.Logger
	cfg   *config.Config
	tomb  utils.Tomb
	http  *http.Client
	srv   *gohttp.Server
	ami   ami.AMI
	batch *batch
	attrs map[string]string
	sig   chan bool
}

// NewInit to activate, success add node info
func NewInit(cfg *config.Config) (*Initialize, error) {
	ops, err := cfg.Init.Cloud.HTTP.ToClientOptions()
	if err != nil {
		return nil, err
	}

	init := &Initialize{
		cfg:   cfg,
		sig:   make(chan bool, 1),
		http:  http.NewClient(ops),
		attrs: map[string]string{},
		log:   log.With(log.Any("core", "Initialize")),
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

	init.ami, err = ami.NewAMI(cfg.Engine)
	if err != nil {
		return nil, err
	}
	return init, nil
}

func (init *Initialize) Start() {
	if init.cfg.Init.ActivateConfig.Server.Listen == "" {
		err := init.tomb.Go(init.activating)
		if err != nil {
			init.log.Error("failed to start report and process routine", log.Error(err))
			return
		}
	} else {
		err := init.tomb.Go(init.startServer)
		if err != nil {
			init.log.Error("init", log.Error(err))
		}
	}
}

func (init *Initialize) Close() {
	if init.srv != nil {
		init.closeServer()
	}
	init.tomb.Kill(nil)
	init.tomb.Wait()
}

func (init *Initialize) WaitAndClose() {
	if _, ok := <-init.sig; !ok {
		init.log.Error("Initialize get sig error")
	}
	init.Close()
}
