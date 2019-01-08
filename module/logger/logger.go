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
	*logrus.Entry
}

// WithFields adds a map of fields to the Entry.
func (e *Entry) WithFields(vs ...string) *Entry {
	fs := logrus.Fields{}
	for index := 0; index < len(vs)-1; index = index + 2 {
		fs[vs[index]] = vs[index+1]
	}
	return &Entry{Entry: e.Entry.WithFields(fs)}
}

// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func (e *Entry) WithError(err error) *Entry {
	return &Entry{Entry: e.Entry.WithError(err)}
}

// Log the globel logger
var Log *Entry

func init() {
	log := logrus.New()
	log.SetReportCaller(true)
	Log = &Entry{Entry: logrus.NewEntry(log)}
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

	Log = WithFields(fields...)
	Log.Entry.Logger.Level = logLevel
	Log.Entry.Logger.Out = logOutWriter
	Log.Entry.Logger.Formatter = newFormatter(c.Format, true)
	if fileHook != nil {
		Log.Entry.Logger.Hooks.Add(fileHook)
	}
	return nil
}

// WithFields adds a map of fields to the Entry.
func WithFields(vs ...string) *Entry {
	fs := logrus.Fields{}
	for index := 0; index < len(vs)-1; index = index + 2 {
		fs[vs[index]] = vs[index+1]
	}
	return &Entry{Entry: Log.Entry.WithFields(fs)}
}

// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func WithError(err error) *Entry {
	return &Entry{Entry: Log.Entry.WithError(err)}
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
