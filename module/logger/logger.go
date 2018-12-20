package logger

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/baidu/openedge/module/config"
	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Entry logging entry
type Entry struct {
	entry *logrus.Entry
}

// WithFields adds a map of fields to the Entry.
func (e *Entry) WithFields(vs ...string) *Entry {
	fs := logrus.Fields{}
	for index := 0; index < len(vs)-1; index = index + 2 {
		fs[vs[index]] = vs[index+1]
	}
	return &Entry{entry: e.entry.WithFields(fs)}
}

// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func (e *Entry) WithError(err error) *Entry {
	return &Entry{entry: e.entry.WithError(err)}
}

// // Debug log debug info
// func (e *Entry) Debug(args ...interface{}) {
// 	e.entry.Debug(args...)
// }

// // Info log info
// func (e *Entry) Info(args ...interface{}) {
// 	e.entry.Info(args...)
// }

// // Warn log warning info
// func (e *Entry) Warn(args ...interface{}) {
// 	e.entry.Warn(args...)
// }

// // Error log error info
// func (e *Entry) Error(args ...interface{}) {
// 	e.entry.Error(args...)
// }

// Debugf log debug info
func (e *Entry) Debugf(format string, args ...interface{}) {
	e.entry.Debugf(format, args...)
}

// Infof log info
func (e *Entry) Infof(format string, args ...interface{}) {
	e.entry.Infof(format, args...)
}

// Warnf log warning info
func (e *Entry) Warnf(format string, args ...interface{}) {
	e.entry.Warnf(format, args...)
}

// Errorf log error info
func (e *Entry) Errorf(format string, args ...interface{}) {
	e.entry.Errorf(format, args...)
}

// Debugln log debug info
func (e *Entry) Debugln(args ...interface{}) {
	e.entry.Debugln(args...)
}

// Infoln log info
func (e *Entry) Infoln(args ...interface{}) {
	e.entry.Infoln(args...)
}

// Warnln log warning info
func (e *Entry) Warnln(args ...interface{}) {
	e.entry.Warnln(args...)
}

// Errorln log error info
func (e *Entry) Errorln(args ...interface{}) {
	e.entry.Errorln(args...)
}

var root = &Entry{entry: logrus.NewEntry(logrus.New())}

// Init init logger
func Init(c config.Logger, fields ...string) error {
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
	root.entry.Logger.Level = logLevel
	root.entry.Logger.Out = logOutWriter
	root.entry.Logger.Formatter = newFormatter(c.Format, true)
	if fileHook != nil {
		root.entry.Logger.Hooks.Add(fileHook)
	}
	return nil
}

// WithFields adds a map of fields to the Entry.
func WithFields(vs ...string) *Entry {
	fs := logrus.Fields{}
	for index := 0; index < len(vs)-1; index = index + 2 {
		fs[vs[index]] = vs[index+1]
	}
	return &Entry{entry: root.entry.WithFields(fs)}
}

// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func WithError(err error) *Entry {
	return &Entry{entry: root.entry.WithError(err)}
}

// // Debug log debug info
// func Debug(args ...interface{}) {
// 	root.entry.Debug(args...)
// }

// // Info log info
// func Info(args ...interface{}) {
// 	root.entry.Info(args...)
// }

// // Warn log warning info
// func Warn(args ...interface{}) {
// 	root.entry.Warn(args...)
// }

// // Error log error info
// func Error(args ...interface{}) {
// 	root.entry.Error(args...)
// }

// Debugf log debug info
func Debugf(format string, args ...interface{}) {
	root.entry.Debugf(format, args...)
}

// Infof log info
func Infof(format string, args ...interface{}) {
	root.entry.Infof(format, args...)
}

// Warnf log warning info
func Warnf(format string, args ...interface{}) {
	root.entry.Warnf(format, args...)
}

// Errorf log error info
func Errorf(format string, args ...interface{}) {
	root.entry.Errorf(format, args...)
}

// Debugln log debug info
func Debugln(args ...interface{}) {
	root.entry.Debugln(args...)
}

// Infoln log info
func Infoln(args ...interface{}) {
	root.entry.Infoln(args...)
}

// Warnln log warning info
func Warnln(args ...interface{}) {
	root.entry.Warnln(args...)
}

// Errorln log error info
func Errorln(args ...interface{}) {
	root.entry.Errorln(args...)
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
