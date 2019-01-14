package sdk

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type logger struct {
	entry *logrus.Entry
}

func (l *logger) WithField(key string, value interface{}) openedge.Logger {
	return &logger{l.entry.WithField(key, value)}
}

func (l *logger) WithError(err error) openedge.Logger {
	return &logger{l.entry.WithError(err)}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logger) Debugln(args ...interface{}) {
	l.entry.Debugln(args...)
}

func (l *logger) Infoln(args ...interface{}) {
	l.entry.Infoln(args...)
}

func (l *logger) Warnln(args ...interface{}) {
	l.entry.Warnln(args...)
}

func (l *logger) Errorln(args ...interface{}) {
	l.entry.Errorln(args...)
}

func (l *logger) Fatalln(args ...interface{}) {
	l.entry.Fatalln(args...)
}

// InitLogger of global logger
func InitLogger(c *openedge.LogInfo, fields ...string) error {
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

	entry := logrus.NewEntry(logrus.New())
	entry.Logger.SetReportCaller(true)
	entry.Logger.Level = logLevel
	entry.Logger.Out = logOutWriter
	entry.Logger.Formatter = newFormatter(c.Format, true)
	if fileHook != nil {
		entry.Logger.Hooks.Add(fileHook)
	}
	logrusFields := logrus.Fields{}
	for index := 0; index < len(fields)-1; index = index + 2 {
		logrusFields[fields[index]] = fields[index+1]
	}
	openedge.SetGlobalLogger(&logger{
		entry: entry.WithFields(logrusFields),
	})
	return nil
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
