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
type Entry interface {
	// Add an error as single field (using the key defined in ErrorKey) to the Entry.
	WithError(err error) Entry
	// Add a single field to the Entry.
	WithField(key string, value interface{}) Entry
	// Entry Printf family functions
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	// Entry Println family functions
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
}

type logrusEntry struct {
	*logrus.Entry
}

// Add an error as single field (using the key defined in ErrorKey) to the Entry.
func (e *logrusEntry) WithError(err error) Entry {
	return &logrusEntry{e.Entry.WithError(err)}
}

// Add a single field to the Entry.
func (e *logrusEntry) WithField(key string, value interface{}) Entry {
	return &logrusEntry{e.Entry.WithField(key, value)}
}

// Log the globel logger
var Log Entry

func init() {
	log := logrus.New()
	// log.SetReportCaller(true)
	Log = &logrusEntry{logrus.NewEntry(log)}
}

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

	log := Log.(*logrusEntry)
	log.Logger.Level = logLevel
	log.Logger.Out = logOutWriter
	log.Logger.Formatter = newFormatter(c.Format, true)
	if fileHook != nil {
		log.Logger.Hooks.Add(fileHook)
	}
	logrusFields := logrus.Fields{}
	for index := 0; index < len(fields)-1; index = index + 2 {
		logrusFields[fields[index]] = fields[index+1]
	}
	Log = &logrusEntry{log.Entry.WithFields(logrusFields)}
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
