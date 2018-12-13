package logger

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// var mutex sync.Mutex
var root = logrus.NewEntry(logrus.New())

// Init init logger
func Init(c Config, fields ...string) error {
	// mutex.Lock()
	// defer mutex.Unlock()

	var logOutWriter io.Writer
	if c.Console == true {
		logOutWriter = os.Stdout
	} else {
		logOutWriter = ioutil.Discard
	}
	logLevel, err := logrus.ParseLevel(c.Level)
	if err != nil {
		logLevel = logrus.DebugLevel
	}

	var fileHook logrus.Hook
	if len(c.Path) != 0 {
		err := os.MkdirAll(filepath.Dir(c.Path), 0755)
		if err != nil {
			return err
		}
		fileHook, err = newFileHook(fileConfig{
			Filename:   c.Path,
			Formatter:  newFormatter(c.Format, false),
			Level:      logLevel,
			MaxAge:     c.Age.Max,  //days
			MaxSize:    c.Size.Max, // megabytes
			MaxBackups: c.Backup.Max,
			Compress:   true,
		})
		if err != nil {
			return err
		}
	}

	root = WithFields(fields...)
	root.Logger.Level = logLevel
	root.Logger.Out = logOutWriter
	root.Logger.Formatter = newFormatter(c.Format, true)
	if fileHook != nil {
		root.Logger.Hooks.Add(fileHook)
	}
	return nil
}

// WithFields adds a map of fields to the Entry.
func WithFields(vs ...string) *logrus.Entry {
	// mutex.Lock()
	// defer mutex.Unlock()
	fs := logrus.Fields{}
	for index := 0; index < len(vs)-1; index = index + 2 {
		fs[vs[index]] = vs[index+1]
	}
	return root.WithFields(fs)
}

// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func WithError(err error) *logrus.Entry {
	return root.WithError(err)
}

// Debug log debug info
func Debug(args ...interface{}) {
	root.Debug(args...)
}

// Info log info
func Info(args ...interface{}) {
	root.Info(args...)
}

// Warn log warning info
func Warn(args ...interface{}) {
	root.Warn(args...)
}

// Error log error info
func Error(args ...interface{}) {
	root.Error(args...)
}

// Debugf log debug info
func Debugf(format string, args ...interface{}) {
	root.Debugf(format, args...)
}

// Infof log info
func Infof(format string, args ...interface{}) {
	root.Infof(format, args...)
}

// Warnf log warning info
func Warnf(format string, args ...interface{}) {
	root.Warnf(format, args...)
}

// Errorf log error info
func Errorf(format string, args ...interface{}) {
	root.Errorf(format, args...)
}

// Debugln log debug info
func Debugln(args ...interface{}) {
	root.Debugln(args...)
}

// Infoln log info
func Infoln(args ...interface{}) {
	root.Infoln(args...)
}

// Warnln log warning info
func Warnln(args ...interface{}) {
	root.Warnln(args...)
}

// Errorln log error info
func Errorln(args ...interface{}) {
	root.Errorln(args...)
}

type fileConfig struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	LocalTime  bool
	Compress   bool
	Level      logrus.Level
	Formatter  logrus.Formatter
}

type fileHook struct {
	config fileConfig
	writer io.Writer
}

func newFileHook(config fileConfig) (logrus.Hook, error) {
	hook := fileHook{
		config: config,
	}

	var zeroLevel logrus.Level
	if hook.config.Level == zeroLevel {
		hook.config.Level = logrus.InfoLevel
	}
	var zeroFormatter logrus.Formatter
	if hook.config.Formatter == zeroFormatter {
		hook.config.Formatter = new(logrus.TextFormatter)
	}

	hook.writer = &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		LocalTime:  config.LocalTime,
		Compress:   config.Compress,
	}

	return &hook, nil
}

// Levels Levels
func (hook *fileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Fire Fire
func (hook *fileHook) Fire(entry *logrus.Entry) (err error) {
	if hook.config.Level < entry.Level {
		return nil
	}
	b, err := hook.config.Formatter.Format(entry)
	if err != nil {
		return err
	}
	hook.writer.Write(b)
	return nil
}

func newFormatter(format string, color bool) logrus.Formatter {
	var formatter logrus.Formatter
	if strings.ToLower(format) == "json" {
		formatter = &logrus.JSONFormatter{}
	} else {
		if runtime.GOOS == "windows" {
			color = false
		}
		formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: !color}
	}
	return formatter
}
