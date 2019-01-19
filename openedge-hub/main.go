package main

import (
	"os"
	"os/signal"
	"syscall"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/openedge-hub/broker"
	"github.com/baidu/openedge/openedge-hub/config"
	"github.com/baidu/openedge/openedge-hub/persist"
	"github.com/baidu/openedge/openedge-hub/rule"
	"github.com/baidu/openedge/openedge-hub/server"
	"github.com/baidu/openedge/openedge-hub/session"
	sdk "github.com/baidu/openedge/sdk/go"
	"github.com/baidu/openedge/utils"
)

// "net/http"
// _ "net/http/pprof"
// "path/filepath"
// "runtime/trace"

type mo struct {
	cfg      config.Config
	Rules    *rule.Manager
	Sessions *session.Manager
	broker   *broker.Broker
	servers  *server.Manager
	factory  *persist.Factory
}

const defaultConfigPath = "etc/openedge/service.yml"

func (m *mo) init() error {
	err := utils.LoadYAML(defaultConfigPath, &m.cfg)
	if err != nil {
		openedge.Errorln("failed to load config:", err.Error())
		return err
	}
	err = sdk.InitLogger(&m.cfg.Logger)
	if err != nil {
		openedge.Errorln("failed to init logger:", err.Error())
		return err
	}
	m.factory, err = persist.NewFactory(m.cfg.Storage.Dir)
	if err != nil {
		openedge.Errorln("failed to new factory:", err.Error())
		return err
	}
	m.broker, err = broker.NewBroker(&m.cfg, m.factory)
	if err != nil {
		openedge.Errorln("failed to new broker:", err.Error())
		return err
	}
	m.Rules, err = rule.NewManager(m.cfg.Subscriptions, m.broker)
	if err != nil {
		openedge.Errorln("failed to new rule manager:", err.Error())
		return err
	}
	m.Sessions, err = session.NewManager(&m.cfg, m.broker.Flow, m.Rules, m.factory)
	if err != nil {
		openedge.Errorln("failed to new session manager:", err.Error())
		return err
	}
	m.servers, err = server.NewManager(m.cfg.Listen, m.cfg.Certificate, m.Sessions.Handle)
	if err != nil {
		openedge.Errorln("failed to new server manager:", err.Error())
		return err
	}
	m.Rules.Start()
	m.servers.Start()
	return nil
}

func (m *mo) close() {
	if m.Rules != nil {
		m.Rules.Close()
	}
	if m.Sessions != nil {
		m.Sessions.Close()
	}
	if m.servers != nil {
		m.servers.Close()
	}
	if m.broker != nil {
		m.broker.Close()
	}
	if m.factory != nil {
		m.factory.Close()
	}
}

func main() {

	// // go tool pprof http://localhost:6060/debug/pprof/profile
	// go func() {
	// 	err := http.ListenAndServe("localhost:6060", nil)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, "Start profile failed: ", err.Error())
	// 		return
	// 	}
	// }()

	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

	var m mo
	err := m.init()
	if err != nil {
		openedge.Fatalln("failed to init openedge-hub:", err.Error())
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
	m.close()
}
