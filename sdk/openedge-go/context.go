package openedge

import (
	fmt "fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
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
	// AppConfFileName application config file name
	AppConfFileName = "application.yml"
	// DefaultSockFile sock file of openedge by default
	DefaultSockFile = "/var/run/openedge.sock"
	// DefaultPidFile pid file of openedge by default
	DefaultPidFile = "/var/run/openedge.pid"
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
	// returns the system configuration of the service, such as hub and logger
	Config() *ServiceConfig
	// loads the custom configuration of the service
	LoadConfig(interface{}) error
	// creates a Client that connects to the Hub through system configuration,
	// you can specify the Client ID and the topic information of the subscription.
	NewHubClient(string, []mqtt.TopicInfo) (*mqtt.Dispatcher, error)
	// returns logger interface
	Log() logger.Logger
	// waiting to exit, receiving SIGTERM and SIGINT signals
	Wait()
	// returns wait channel
	WaitChan() <-chan os.Signal

	// Master RESTful API

	// updates system and
	UpdateSystem(string, bool) error
	// inspects system stats
	InspectSystem() (*Inspect, error)
	// gets an available port of the host
	GetAvailablePort() (string, error)
	// starts an instance of a service
	StartServiceInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
	// stop an instance of a service
	StopServiceInstance(serviceName, instanceName string) error
}

type ctx struct {
	*Client
	cfg ServiceConfig
	log logger.Logger
}

func (c *ctx) NewHubClient(cid string, subs []mqtt.TopicInfo) (*mqtt.Dispatcher, error) {
	if c.cfg.Hub.Address == "" {
		return nil, fmt.Errorf("hub not configured")
	}
	cc := c.cfg.Hub
	if cid != "" {
		cc.ClientID = cid
	}
	if subs != nil {
		cc.Subscriptions = subs
	}
	return mqtt.NewDispatcher(cc, c.log.WithField("cid", cid)), nil
}

func (c *ctx) LoadConfig(cfg interface{}) error {
	return utils.LoadYAML(DefaultConfFile, cfg)
}

func (c *ctx) Config() *ServiceConfig {
	return &c.cfg
}

func (c *ctx) Log() logger.Logger {
	return c.log
}

func (c *ctx) Wait() {
	<-c.WaitChan()
}

func (c *ctx) WaitChan() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	return sig
}

func newContext() (*ctx, error) {
	var cfg ServiceConfig
	err := utils.LoadYAML(DefaultConfFile, &cfg)
	if err != nil {
		return nil, err
	}
	name, ok := os.LookupEnv(EnvServiceNameKey)
	if !ok {
		name = "<unknown>"
	}
	log, err := logger.InitLogger(&cfg.Logger, "service", name)
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
