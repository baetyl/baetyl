package main

import (
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/openedge-hub/broker"
	"github.com/baidu/openedge/openedge-hub/config"
	"github.com/baidu/openedge/openedge-hub/persist"
	"github.com/baidu/openedge/openedge-hub/rule"
	"github.com/baidu/openedge/openedge-hub/server"
	"github.com/baidu/openedge/openedge-hub/session"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// "net/http"
// _ "net/http/pprof"
// "path/filepath"
// "runtime/trace"

type mo struct {
	ctx      openedge.Context
	cfg      config.Config
	Rules    *rule.Manager
	Sessions *session.Manager
	broker   *broker.Broker
	servers  *server.Manager
	factory  *persist.Factory
	log      logger.Logger
}

func (m *mo) start() error {
	err := m.ctx.LoadConfig(&m.cfg)
	if err != nil {
		m.log.Errorln("failed to load config:", err.Error())
		return err
	}
	m.factory, err = persist.NewFactory(m.cfg.Storage.Dir)
	if err != nil {
		m.log.Errorln("failed to new factory:", err.Error())
		return err
	}
	m.broker, err = broker.NewBroker(&m.cfg, m.factory)
	if err != nil {
		m.log.Errorln("failed to new broker:", err.Error())
		return err
	}
	m.Rules, err = rule.NewManager(m.cfg.Subscriptions, m.broker)
	if err != nil {
		m.log.Errorln("failed to new rule manager:", err.Error())
		return err
	}
	m.Sessions, err = session.NewManager(&m.cfg, m.broker.Flow, m.Rules, m.factory)
	if err != nil {
		m.log.Errorln("failed to new session manager:", err.Error())
		return err
	}
	m.servers, err = server.NewManager(m.cfg.Listen, m.cfg.Certificate, m.Sessions.Handle)
	if err != nil {
		m.log.Errorln("failed to new server manager:", err.Error())
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

	openedge.Run(func(ctx openedge.Context) error {
		m := mo{ctx: ctx, log: ctx.Log()}
		defer m.close()
		err := m.start()
		if err != nil {
			return err
		}
		ctx.Wait()
		return nil
	})
}
