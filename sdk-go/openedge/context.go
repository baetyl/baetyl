package openedge

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/utils"
)

// Env variable keys
const (
	EnvHostOSKey         = "OPENEDGE_HOST_OS"
	EnvMasterAPIKey      = "OPENEDGE_MASTER_API"
	EnvRunningModeKey    = "OPENEDGE_RUNNING_MODE"
	EnvServiceNameKey    = "OPENEDGE_SERVICE_NAME"
	EnvServiceTokenKey   = "OPENEDGE_SERVICE_TOKEN"
	EnvServiceAddressKey = "OPENEDGE_SERVICE_ADDRESS"
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

// Context of service
type Context interface {
	Config() *ServiceConfig
	UpdateSystem([]byte) error
	InspectSystem() (*Inspect, error)
	Log() logger.Logger
	Wait()

	GetAvailablePort() (string, error)
	// GetServiceInfo(serviceName string) (*ServiceInfo, error)
	StartServiceInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
	StopServiceInstance(serviceName, instanceName string) error
}

type ctx struct {
	*Client
	cfg ServiceConfig
	log logger.Logger
}

func (c *ctx) Config() *ServiceConfig {
	return &c.cfg
}

func (c *ctx) Log() logger.Logger {
	return c.log
}

func (c *ctx) Wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}

func newContext() (*ctx, error) {
	var cfg ServiceConfig
	err := utils.LoadYAML(DefaultConfFile, &cfg)
	if err != nil {
		return nil, err
	}
	log, err := logger.InitLogger(&cfg.Logger, "service", cfg.Name)
	if err != nil {
		return nil, err
	}

	c := &ctx{
		cfg: cfg,
		log: log,
	}
	c.Client, err = NewEnvClient()
	if err != nil {
		log.Warnln(err.Error())
	}
	return c, nil
}
