package baetyl

import (
	fmt "fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"
)

// Mode keys
const (
	ModeNative = "native"
	ModeDocker = "docker"
)

// OTA types
const (
	OTAAPP = "APP"
	OTAMST = "MST"
)

// OTA steps
const (
	OTAKeyStep  = "step"
	OTAKeyType  = "type"
	OTAKeyTrace = "trace"

	OTAReceived    = "RECEIVED"    // [agent] ota event is received
	OTAUpdating    = "UPDATING"    // [master] to update app or master
	OTAUpdated     = "UPDATED"     // [master][finished] app or master is updated
	OTARestarting  = "RESTARTING"  // [master] to restart master
	OTARestarted   = "RESTARTED"   // [master] master is restarted
	OTARollingBack = "ROLLINGBACK" // [master] to roll back app or master
	OTARolledBack  = "ROLLEDBACK"  // [master][finished] app or master is rolled back
	OTAFailure     = "FAILURE"     // [master/agent][finished] failed to update app or master
	OTATimeout     = "TIMEOUT"     // [agent][finished] ota is timed out
)

// CheckOK print OK if binary is valid
const CheckOK = "OK!"

// Env keys
const (
	EnvHostOSKey                 = "OPENEDGE_HOST_OS"
	EnvMasterAPIKey              = "OPENEDGE_MASTER_API"
	EnvMasterAPIVersionKey       = "OPENEDGE_MASTER_API_VERSION"
	EnvRunningModeKey            = "OPENEDGE_RUNNING_MODE"
	EnvServiceNameKey            = "OPENEDGE_SERVICE_NAME"
	EnvServiceTokenKey           = "OPENEDGE_SERVICE_TOKEN"
	EnvServiceAddressKey         = "OPENEDGE_SERVICE_ADDRESS" // deprecated
	EnvServiceInstanceNameKey    = "OPENEDGE_SERVICE_INSTANCE_NAME"
	EnvServiceInstanceAddressKey = "OPENEDGE_SERVICE_INSTANCE_ADDRESS"

	EnvHostID           = "OPENEDGE_HOST_ID"
	EnvMasterHostSocket = "OPENEDGE_MASTER_HOST_SOCKET"
)

// Path keys
const (
	// AppConfFileName application config file name
	AppConfFileName = "application.yml"
	// AppBackupFileName application backup configuration file
	AppBackupFileName = "application.yml.old"
	// AppStatsFileName application stats file name
	AppStatsFileName = "application.stats"

	// BinFile the file path of master binary
	DefaultBinFile = "bin/baetyl"
	// DefaultBinBackupFile the backup file path of master binary
	DefaultBinBackupFile = "bin/baetyl.old"
	// DefaultSockFile sock file of baetyl by default
	DefaultSockFile = "var/run/baetyl.sock"
	// DefaultConfFile config path of the service by default
	DefaultConfFile = "etc/baetyl/service.yml"
	// DefaultDBDir db dir of the service by default
	DefaultDBDir = "var/db/baetyl"
	// DefaultRunDir  run dir of the service by default
	DefaultRunDir = "var/run/baetyl"
	// DefaultLogDir  log dir of the service by default
	DefaultLogDir = "var/log/baetyl"
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
	// check running mode
	IsNative() bool
	// waiting to exit, receiving SIGTERM and SIGINT signals
	Wait()
	// returns wait channel
	WaitChan() <-chan os.Signal

	// Master RESTful API

	// updates application or master
	UpdateSystem(trace, tp, path string) error
	// inspects system stats
	InspectSystem() (*Inspect, error)
	// gets an available port of the host
	GetAvailablePort() (string, error)
	// reports the stats of the instance of the service
	ReportInstance(stats map[string]interface{}) error
	// starts an instance of the service
	StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
	// stop the instance of the service
	StopInstance(serviceName, instanceName string) error
}

type ctx struct {
	sn  string // service name
	in  string // instance name
	md  string // running mode
	cli *Client
	cfg ServiceConfig
	log logger.Logger
}

func newContext() (*ctx, error) {
	var cfg ServiceConfig
	sn := os.Getenv(EnvServiceNameKey)
	in := os.Getenv(EnvServiceInstanceNameKey)
	md := os.Getenv(EnvRunningModeKey)
	err := utils.LoadYAML(DefaultConfFile, &cfg)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to load config: %s\n", sn, in, err.Error())
	}
	log := logger.InitLogger(cfg.Logger, "service", sn, "instance", in)
	cli, err := NewEnvClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s][%s] failed to create master client: %s\n", sn, in, err.Error())
		log.WithError(err).Errorf("failed to create master client")
	}
	return &ctx{
		sn:  sn,
		in:  in,
		md:  md,
		cfg: cfg,
		cli: cli,
		log: log,
	}, nil
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

func (c *ctx) IsNative() bool {
	return c.md == ModeNative
}

func (c *ctx) WaitChan() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	return sig
}

// InspectSystem inspect all stats
func (c *ctx) InspectSystem() (*Inspect, error) {
	return c.cli.InspectSystem()
}

// UpdateSystem updates and reloads config
func (c *ctx) UpdateSystem(trace, tp, path string) error {
	return c.cli.UpdateSystem(trace, tp, path)
}

// GetAvailablePort gets available port
func (c *ctx) GetAvailablePort() (string, error) {
	return c.cli.GetAvailablePort()
}

// ReportInstance reports the stats of the instance of the service
func (c *ctx) ReportInstance(stats map[string]interface{}) error {
	return c.cli.ReportInstance(c.sn, c.in, stats)
}

// StartInstance starts a new service instance with dynamic config
func (c *ctx) StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	return c.cli.StartInstance(serviceName, instanceName, dynamicConfig)
}

// StopInstance stops a service instance
func (c *ctx) StopInstance(serviceName, instanceName string) error {
	return c.cli.StopInstance(serviceName, instanceName)
}
