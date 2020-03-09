package main

import (
	"fmt"
	"github.com/baetyl/baetyl-core/agent"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/kube"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl/utils"
	"os"
	"os/signal"
	"syscall"
)

type core struct {
	a       agent.Agent
	kubeCli *kube.Client
	driver  store.Driver
	cfg     config.Config
}

func NewCore(cfg config.Config) (*core, error) {
	kubeCli, err := kube.NewClient(cfg.APIServer)
	logger, err := log.Init(cfg.Logger)
	if err != nil {
		return nil, err
	}
	driver := store.NewConfigMapDriver(kubeCli.CoreV1.ConfigMaps(kubeCli.Namespace))
	a, err := agent.NewAgent(cfg.Agent, kubeCli.AppV1.Deployments(kubeCli.Namespace), logger)
	if err != nil {
		return nil, err
	}
	return &core{
		kubeCli: kubeCli,
		driver:  driver,
		cfg:     cfg,
		a:       a,
	}, nil
}

func (c *core) Start() error {
	go c.a.Start()
	return nil
}

func (c *core) Stop() {
}

func main() {
	var cfg config.Config
	err := utils.LoadYAML(common.DefaultConfFile, &cfg)
	if err != nil {
		log.With(log.Any("core", "main")).Error("failed to load config file", log.Error(err))
		os.Exit(1)
	}
	c, err := NewCore(cfg)
	if err != nil {
		log.With(log.Any("core", "main")).Error("failed to create core", log.Error(err))
		os.Exit(1)
	}
	defer c.Stop()
	err = c.Start()
	if err != nil {
		log.With(log.Any("core", "main")).Error("failed to start core", log.Error(err))
		os.Exit(1)
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT,
		syscall.SIGILL, syscall.SIGHUP, syscall.SIGTERM,
		syscall.SIGTRAP, syscall.SIGABRT)
	<-sig
	fmt.Printf("core stopped")
}
