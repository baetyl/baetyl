package openedge

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/utils"
)

// Enviroment variable keys
const (
	EnvHostOSKey       = "OPENEDGE_HOST_OS"
	EnvMasterAPIKey    = "OPENEDGE_MASTER_API"
	EnvServiceModeKey  = "OPENEDGE_SERVICE_MODE"
	EnvServiceNameKey  = "OPENEDGE_SERVICE_NAME"
	EnvServiceTokenKey = "OPENEDGE_SERVICE_TOKEN"
)

const (
	// DefaultConfFile config path of the service by default
	DefaultConfFile = "etc/openedge/service.yml"
	// DefaultDBDir db dir of the service by default
	DefaultDBDir = "var/db/openedge"
	// DefaultRunDir  run dir of the service by default
	DefaultRunDir = "var/run/openedge"
	// DefaultLogDir  log dir of the service by default
	DefaultLogDir = "var/log/openedge"
)

// Context of module
type Context interface {
	Config() *ServiceConfig
	UpdateSystem(*AppConfig) error
	InspectSystem() (*Inspect, error)
	Log() logger.Logger
	Wait()
}

type context struct {
	*Client
	cfg ServiceConfig
	log logger.Logger
}

func (c *context) Config() *ServiceConfig {
	return &c.cfg
}

func (c *context) Log() logger.Logger {
	return c.log
}

func (c *context) Wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}

func newContext() (*context, error) {
	var cfg ServiceConfig
	err := utils.LoadYAML(DefaultConfFile, &cfg)
	if err != nil {
		return nil, err
	}
	log, err := logger.InitLogger(&cfg.Logger, "service", cfg.Name)
	if err != nil {
		return nil, err
	}

	c := &context{
		cfg: cfg,
		log: log,
	}
	c.Client, err = NewEnvClient()
	if err != nil {
		log.Warnln(err.Error())
	}
	return c, nil
}
