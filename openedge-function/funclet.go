package main

import (
	"encoding/json"
	fmt "fmt"
	"os"
	"path"
	"sync"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/function/runtime"
	"github.com/baidu/openedge/module/logger"
	"github.com/docker/distribution/uuid"
)

// funclet is an instance of function
type funclet struct {
	id   string
	cfg  config.Function
	rtc  *runtime.Client
	man  *Manager
	log  logger.Entry
	once sync.Once
}

func (fl *funclet) start() error {
	host, port := fl.id, 50051
	if os.Getenv("OPENEDGE_MODULE_MODE") != "docker" {
		var err error
		host = "127.0.0.1"
		port, err = fl.man.cli.GetPortAvailable(host)
		if err != nil {
			return err
		}
	}
	rc := config.Runtime{}
	rc.Name = fl.id
	rc.Function = fl.cfg
	rc.Server.Address = fmt.Sprintf("%s:%d", host, port)
	rc.Server.Timeout = fl.cfg.Instance.Timeout
	rc.Server.Message.Length.Max = fl.cfg.Instance.Message.Length.Max
	rc.Logger = fl.man.cfg.Logger
	if rc.Logger.Path != "" {
		rc.Logger.Path = path.Join("var", "log", "openedge", fl.cfg.ID, fl.cfg.Name+".log")
	}
	rcd, err := json.Marshal(rc)
	if err != nil {
		return err
	}

	mc := config.Module{}
	mc.Name = fl.cfg.ID
	mc.Alias = fl.id
	mc.Entry = fl.cfg.Entry
	mc.Env = fl.cfg.Env
	mc.Params = []string{"-c", string(rcd)}
	mc.Resources = fl.cfg.Instance.Resources

	fl.log.Debugln("Runtime config:", mc)
	err = fl.man.cli.StartModule(&mc)
	if err != nil {
		return err
	}

	cc := config.NewRuntimeClient(fmt.Sprintf("%s:%d", host, port))
	cc.RuntimeServer = rc.Server
	cc.Backoff.Max = rc.Server.Timeout
	fl.rtc, err = runtime.NewClient(cc)
	if err != nil {
		fl.log.WithError(err).Errorf("failed to create runtime client")
	}
	return err
}

func (fl *funclet) Close() {
	fl.once.Do(func() {
		if fl.rtc != nil {
			fl.rtc.Close()
		}
		mc := &config.Module{Name: fl.cfg.ID, Alias: fl.id}
		err := fl.man.cli.StopModule(mc)
		if err != nil {
			fl.log.WithError(err).Warnf("failed to stop function instance")
		}
	})
}

func (fl *funclet) handle(msg *runtime.Message) (*runtime.Message, error) {
	msg.FunctionName = fl.cfg.Name
	if msg.FunctionInvokeID == "" {
		msg.FunctionInvokeID = uuid.Generate().String()
	}
	return fl.rtc.Handle(msg)
}
