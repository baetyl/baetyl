package openedge

// Enviroment variable keys
const (
	MasterAPIKey   = "OPENEDGE_MASTER_API"
	HostOSKey      = "OPENEDGE_HOST_OS"
	ModuleModeKey  = "OPENEDGE_MODULE_MODE"
	ModuleTokenKey = "OPENEDGE_MODULE_TOKEN"
)

// Context of module
type Context interface {
	Config() *Config
	WaitExit()
	Subscribe(topic TopicInfo, handler func(*Message) error) error
	SendMessage(message *Message) error
	StartService(name string, info *ServiceInfo, config []byte) error
	StopService(name string) error
	UpdateSystem(configPath string) error
}

// master rpc call names
const (
	CallStartService = "openedge.StartService"
)

// StartServiceRequest data
type StartServiceRequest struct {
	Name   string
	Info   ServiceInfo
	Config []byte
}

// StartServiceResponse data
type StartServiceResponse string

// Logger of module
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

var gLogger Logger

func init() {
	gLogger = &fakeLogger{}
}

// GlobalLogger of openedge
func GlobalLogger() Logger {
	return gLogger
}

// SetGlobalLogger for openedge
func SetGlobalLogger(logger Logger) {
	gLogger = logger
}

// WithField to global logger
func WithField(key string, value interface{}) Logger {
	return GlobalLogger().WithField(key, value)
}

// WithError to global logger
func WithError(err error) Logger {
	return GlobalLogger().WithError(err)
}

// Debugf to global logger
func Debugf(format string, args ...interface{}) {
	GlobalLogger().Debugf(format, args...)
}

// Infof to global logger
func Infof(format string, args ...interface{}) {
	GlobalLogger().Infof(format, args...)
}

// Warnf to global logger
func Warnf(format string, args ...interface{}) {
	GlobalLogger().Warnf(format, args...)
}

// Errorf to global logger
func Errorf(format string, args ...interface{}) {
	GlobalLogger().Errorf(format, args...)
}

// Fatalf to global logger
func Fatalf(format string, args ...interface{}) {
	GlobalLogger().Fatalf(format, args...)
}

// Debugln to global logger
func Debugln(args ...interface{}) {
	GlobalLogger().Debugln(args...)
}

// Infoln to global logger
func Infoln(args ...interface{}) {
	GlobalLogger().Infoln(args...)
}

// Warnln to global logger
func Warnln(args ...interface{}) {
	GlobalLogger().Warnln(args...)
}

// Errorln to global logger
func Errorln(args ...interface{}) {
	GlobalLogger().Errorln(args...)
}

// Fatalln to global logger
func Fatalln(args ...interface{}) {
	GlobalLogger().Fatalln(args)
}
