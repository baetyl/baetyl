package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

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

type logger struct {
	entry *logrus.Entry
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{l.entry.WithField(key, value)}
}

func (l *logger) WithError(err error) Logger {
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

// New create a new logger
func New(c LogInfo, fields ...string) Logger {
	logLevel, err := logrus.ParseLevel(c.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level (%s), use default level (info)", c.Level)
		logLevel = logrus.InfoLevel
	}

	var fileHook logrus.Hook
	if c.Path != "" {
		err = os.MkdirAll(filepath.Dir(c.Path), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log directory: %s", err.Error())
		} else {
			fileHook, err = newFileHook(fileConfig{
				Filename:   c.Path,
				Formatter:  newFormatter(c.Format),
				Level:      logLevel,
				MaxAge:     c.Age.Max,  //days
				MaxSize:    c.Size.Max, // megabytes
				MaxBackups: c.Backup.Max,
				Compress:   true,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create log file hook: %s", err.Error())
			}
		}
	}

	entry := logrus.NewEntry(logrus.New())
	entry.Level = logLevel
	entry.Logger.Level = logLevel
	entry.Logger.Formatter = newFormatter(c.Format)
	if fileHook != nil {
		entry.Logger.Hooks.Add(fileHook)
	}
	logrusFields := logrus.Fields{}
	for index := 0; index < len(fields)-1; index = index + 2 {
		logrusFields[fields[index]] = fields[index+1]
	}
	return &logger{entry.WithFields(logrusFields)}
}

// InitLogger init global logger
func InitLogger(c LogInfo, fields ...string) Logger {
	Global = New(c, fields...)
	return Global
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

func newFormatter(format string) logrus.Formatter {
	var formatter logrus.Formatter
	if strings.ToLower(format) == "json" {
		formatter = &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano}
	} else {
		formatter = &logrus.TextFormatter{TimestampFormat: time.RFC3339Nano, FullTimestamp: true, DisableColors: true}
	}
	return formatter
}
