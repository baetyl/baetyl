package initz

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	gohttp "net/http"

	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
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
		return nil, errors.Trace(err)
	}

	ca, err := ioutil.ReadFile(cfg.Init.Cloud.HTTP.CA)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	ops.TLSConfig = &tls.Config{RootCAs: pool}
	ops.TLSConfig.InsecureSkipVerify = cfg.Init.Cloud.HTTP.InsecureSkipVerify

	init := &Initialize{
		cfg:   cfg,
		sig:   make(chan bool, 1),
		http:  http.NewClient(ops),
		attrs: map[string]string{},
		log:   log.With(log.Any("init", "active")),
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

	init.ami, err = ami.NewAMI(cfg.Engine.AmiConfig)
	if err != nil {
		return nil, errors.Trace(err)
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
