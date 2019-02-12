package logger

var gLogger Logger

func init() {
	gLogger = &stdLogger{}
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
