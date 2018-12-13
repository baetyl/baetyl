package function

import (
	"encoding/json"
	fmt "fmt"
	"os"
	"sync"

	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/runtime"
	"github.com/docker/distribution/uuid"
	"github.com/sirupsen/logrus"
)

// funclet is an instance of function
type funclet struct {
	id   string
	cfg  FunctionConfig
	rtc  *runtime.Client
	man  *Manager
	log  *logrus.Entry
	once sync.Once
}

func (fl *funclet) start() error {
	host, port := fl.id, 50051
	if os.Getenv("OPENEDGE_MODULE_MODE") != "docker" {
		var err error
		host = "127.0.0.1"
		port, err = fl.man.api.GetPortAvailable(host)
		if err != nil {
			return err
		}
	}
	rc := runtime.Config{}
	rc.Name = fl.id
	rc.Server.Address = fmt.Sprintf("%s:%d", host, port)
	rc.Server.Timeout = fl.cfg.Instance.Timeout
	rc.Server.Message.Length.Max = fl.cfg.Instance.Message.Length.Max
	rc.Function.Name = fl.cfg.Name
	rc.Function.Handler = fl.cfg.Handler
	rc.Function.CodeDir = fl.cfg.CodeDir
	rc.Logger = fl.man.cfg.Logger
	if rc.Logger.Path != "" {
		rc.Logger.Path = rc.Logger.Path + "." + fl.cfg.Name
	}
	rcd, err := json.Marshal(rc)
	if err != nil {
		return err
	}

	mc := config.Module{}
	mc.Name = fl.id
	mc.Entry = fl.cfg.Entry
	mc.Env = fl.cfg.Env
	mc.Params = []string{"-c", string(rcd)}
	mc.Resources = fl.cfg.Instance.Resources

	fl.log.Debug("Runtime config", mc)
	err = fl.man.api.StartModule(&mc)
	if err != nil {
		return err
	}

	cc := runtime.NewClientConfig(fmt.Sprintf("%s:%d", host, port))
	cc.ServerConfig = rc.Server
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
		err := fl.man.api.StopModule(fl.id)
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
