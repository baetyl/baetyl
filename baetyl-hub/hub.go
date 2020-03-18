package hub

import (
	"path/filepath"

	"github.com/baetyl/baetyl/baetyl-hub/config"
	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

const configDir = "/etc/baetyl-hub"
const configFile = "config.yml"

type hub struct {
	ctx     baetyl.Context
	cfg     config.Config
	log     logger.Logger
	rules   *ruleManager
	sess    *sessionManager
	broker  *broker
	server  *server
	storage *storage
}

func Run(ctx baetyl.Context) error {
	h := hub{ctx: ctx, log: ctx.Log()}

	err := h.ctx.LoadConfig(filepath.Join(configDir, configFile), &h.cfg)
	if err != nil {
		h.log.Errorln("failed to load config:", err.Error())
		return err
	}

	err = h.startStorage()
	if err != nil {
		h.log.Errorln("failed to new storage:", err.Error())
		return err
	}
	defer h.stopStorage()

	err = h.startBroker()
	if err != nil {
		h.log.Errorln("failed to new broker:", err.Error())
		return err
	}
	defer h.stopBroker()

	err = h.startRules()
	if err != nil {
		h.log.Errorln("failed to new rule manager:", err.Error())
		return err
	}
	defer h.stopRules()

	err = h.startSession()
	if err != nil {
		h.log.Errorln("failed to new session manager:", err.Error())
		return err
	}
	defer h.stopSession()

	err = h.runServer()
	if err != nil {
		h.log.Errorln("failed to run server manager:", err.Error())
		return err
	}
	defer h.stopServer()

	<-ctx.Done()
	return nil
}
