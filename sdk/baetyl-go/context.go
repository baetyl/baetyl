package baetyl

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"
)

//go:generate mockgen -destination=mock/context.go -package=baetyl github.com/baetyl/baetyl/sdk/baetyl-go Context

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
	// deprecated
	EnvHostID                    = "OPENEDGE_HOST_ID"
	EnvHostOSKey                 = "OPENEDGE_HOST_OS"
	EnvMasterAPIKey              = "OPENEDGE_MASTER_API"
	EnvMasterAPIVersionKey       = "OPENEDGE_MASTER_API_VERSION"
	EnvRunningModeKey            = "OPENEDGE_RUNNING_MODE"
	EnvServiceNameKey            = "OPENEDGE_SERVICE_NAME"
	EnvServiceTokenKey           = "OPENEDGE_SERVICE_TOKEN"
	EnvServiceAddressKey         = "OPENEDGE_SERVICE_ADDRESS" // deprecated
	EnvServiceInstanceNameKey    = "OPENEDGE_SERVICE_INSTANCE_NAME"
	EnvServiceInstanceAddressKey = "OPENEDGE_SERVICE_INSTANCE_ADDRESS"

	// new envs
	EnvKeyHostID                 = "BAETYL_HOST_ID"
	EnvKeyHostOS                 = "BAETYL_HOST_OS"
	EnvKeyHostSN                 = "BAETYL_HOST_SN"
	EnvKeyMasterAPISocket        = "BAETYL_MASTER_API_SOCKET"
	EnvKeyMasterAPIAddress       = "BAETYL_MASTER_API_ADDRESS"
	EnvKeyMasterAPIVersion       = "BAETYL_MASTER_API_VERSION"
	EnvKeyServiceMode            = "BAETYL_SERVICE_MODE"
	EnvKeyServiceName            = "BAETYL_SERVICE_NAME"
	EnvKeyServiceToken           = "BAETYL_SERVICE_TOKEN"
	EnvKeyServiceInstanceName    = "BAETYL_SERVICE_INSTANCE_NAME"
	EnvKeyServiceInstanceAddress = "BAETYL_SERVICE_INSTANCE_ADDRESS"
)

// Path keys
const (
	// AppConfFileName application config file name
	AppConfFileName = "application.yml"
	// AppBackupFileName application backup configuration file
	AppBackupFileName = "application.yml.old"
	// AppStatsFileName application stats file name
	AppStatsFileName = "application.stats"
	// MetadataFileName application metadata file name
	MetadataFileName = "metadata.yml"

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
	// DefaultMasterConfDir master config dir by default
	DefaultMasterConfDir = "etc/baetyl"
	// DefaultMasterConfFile master config file by default
	DefaultMasterConfFile = "etc/baetyl/conf.yml"

	// backward compatibility
	// PreviousDBDir previous db dir of the service
	PreviousDBDir = "var/db/openedge"
	// PreviousMasterConfDir previous master config dir
	PreviousMasterConfDir = "etc/openedge"
	// PreviousMasterConfFile previous master config file
	PreviousMasterConfFile = "etc/openedge/openedge.yml"
	// PreviousBinBackupFile the backup file path of master binary
	PreviousBinBackupFile = "bin/openedge.old"
	// PreviousLogDir  log dir of the service by default
	PreviousLogDir = "var/log/openedge"
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
	cfg ServiceConfig
	log logger.Logger
	*Client
}

func newContext() (*ctx, error) {
	var cfg ServiceConfig
	md := os.Getenv(EnvKeyServiceMode)
	sn := os.Getenv(EnvKeyServiceName)
	in := os.Getenv(EnvKeyServiceInstanceName)
	if md == "" {
		md = os.Getenv(EnvRunningModeKey)
		sn = os.Getenv(EnvServiceNameKey)
		in = os.Getenv(EnvServiceInstanceNameKey)
	}

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
		sn:     sn,
		in:     in,
		md:     md,
		cfg:    cfg,
		log:    log,
		Client: cli,
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

func (c *ctx) ReportInstance(stats map[string]interface{}) error {
	return c.Client.ReportInstance(c.sn, c.in, stats)
}
